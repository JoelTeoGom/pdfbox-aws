package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"pdf-box-aws/internal/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDbRepo struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBStore(ctx context.Context, tableName string) *DynamoDbRepo {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	return &DynamoDbRepo{
		client:    client,
		tableName: tableName,
	}
}

func (r *DynamoDbRepo) Get(ctx context.Context, userID, fileID string) (*domain.File, error) {
	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "FILE#" + fileID},
		},
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	})
	if err != nil {
		return nil, fmt.Errorf("dynamo get: %w", err)
	}
	if out.Item == nil {
		return nil, domain.ErrFileNotFound // a miss is NOT an error in Dynamo
	}

	var fi fileItem
	attributevalue.UnmarshalMap(out.Item, &fi)

	return fi.ToDomainFile(), nil
}

func (r *DynamoDbRepo) Save(ctx context.Context, file *domain.File) error {
	if file == nil {
		return errors.New("file is nil")
	}
	item, err := attributevalue.MarshalMap(FromDomainFile(file))
	if err != nil {
		return fmt.Errorf("dynamo marshal: %w", err)
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
		// PutItem overwrites by default. Refuse to clobber an existing item so
		// an ID collision surfaces as an error instead of losing the old file.
		ConditionExpression:    aws.String("attribute_not_exists(PK)"),
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	})
	if err != nil {
		var conditionFailed *types.ConditionalCheckFailedException
		if errors.As(err, &conditionFailed) {
			return domain.ErrFileAlreadyExists
		}
		return fmt.Errorf("dynamo save: %w", err)
	}
	return nil
}

// maxListPages bounds how many Dynamo pages one List call walks. The status
// filter is applied after Limit, so a user whose most recent files are all
// deleted needs several pages to fill a single response; the cap keeps that
// from turning one request into an unbounded walk of the partition.
const maxListPages = 5

func (r *DynamoDbRepo) List(ctx context.Context, userID, cursor string, limit int) ([]*domain.File, string, error) {
	var startKey map[string]types.AttributeValue
	if cursor != "" {
		decodedKey, err := DecodeCursor(cursor)
		if err != nil {
			// The cursor is echoed back by the client, so a broken one is a bad
			// request rather than a fault on our side.
			return nil, "", fmt.Errorf("%w: decode cursor: %v", domain.ErrInvalidInput, err)
		}
		startKey = decodedKey
	}

	files := make([]*domain.File, 0, limit)

	for page := 0; len(files) < limit && page < maxListPages; page++ {
		out, err := r.client.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(r.tableName),
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :prefix)"),
			// Deleted and rejected files stay in the table as a record of what
			// happened, but they are not part of what the owner browses.
			FilterExpression:         aws.String("#status <> :deleted AND #status <> :rejected"),
			ExpressionAttributeNames: map[string]string{"#status": "status"},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":       &types.AttributeValueMemberS{Value: "USER#" + userID},
				":prefix":   &types.AttributeValueMemberS{Value: "FILE#"},
				":deleted":  &types.AttributeValueMemberS{Value: string(domain.StatusDeleted)},
				":rejected": &types.AttributeValueMemberS{Value: string(domain.StatusRejected)},
			},
			// Limit counts items read before filtering, so ask only for what is
			// still missing: filtering can shrink a page but never grow it.
			Limit:                  aws.Int32(int32(limit - len(files))),
			ExclusiveStartKey:      startKey,
			ScanIndexForward:       aws.Bool(false), // newest first, by ULID
			ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
		})
		if err != nil {
			return nil, "", fmt.Errorf("dynamo query: %w", err)
		}

		for _, item := range out.Items {
			var storedFile fileItem
			if err := attributevalue.UnmarshalMap(item, &storedFile); err != nil {
				return nil, "", fmt.Errorf("dynamo unmarshal: %w", err)
			}
			files = append(files, storedFile.ToDomainFile())
		}

		startKey = out.LastEvaluatedKey
		if len(startKey) == 0 {
			return files, "", nil // partition exhausted, there is no next page
		}
	}

	return files, EncodeCursor(startKey), nil
}

func (r *DynamoDbRepo) UpdateStatus(ctx context.Context, userID, fileID string, status domain.Status) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "FILE#" + fileID},
		},
		UpdateExpression: aws.String("SET #status = :status"),
		// UpdateItem is an upsert: without this guard a status change for an
		// unknown ID would create the item rather than fail, which is not how
		// an SQL UPDATE on a missing row behaves.
		ConditionExpression: aws.String("attribute_exists(PK)"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: string(status)},
		},
	})
	if err != nil {
		var conditionFailed *types.ConditionalCheckFailedException
		if errors.As(err, &conditionFailed) {
			return domain.ErrFileNotFound
		}
		return fmt.Errorf("dynamo update status: %w", err)
	}
	return nil
}
