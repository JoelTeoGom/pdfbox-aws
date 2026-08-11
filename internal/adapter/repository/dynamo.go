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

func (r *DynamoDbRepo) List(ctx context.Context, userID, cursor string, limit int) ([]*domain.File, string, error) {
	in := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: "USER#" + userID},
			":prefix": &types.AttributeValueMemberS{Value: "FILE#"},
		},
		Limit:                  aws.Int32(int32(limit)),
		ScanIndexForward:       aws.Bool(false), // newest first
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}

	if cursor != "" {
		if decodedKey, err := DecodeCursor(cursor); err != nil {
			return nil, "", fmt.Errorf("decode cursor: %w", err)
		} else {
			in.ExclusiveStartKey = decodedKey
		}
	}

	out, err := r.client.Query(ctx, in)
	if err != nil {
		return nil, "", fmt.Errorf("dynamo query: %w", err)
	}
	next := EncodeCursor(out.LastEvaluatedKey) // "" when there are no more pages

	files := make([]*domain.File, 0, len(out.Items))
	for _, item := range out.Items {
		var fi fileItem
		attributevalue.UnmarshalMap(item, &fi)
		files = append(files, fi.ToDomainFile())
	}

	return files, next, nil
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
