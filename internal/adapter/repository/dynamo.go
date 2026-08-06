package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"pdf-box-aws/internal/adapter/repository"
	"pdf-box-aws/internal/domain"
	"pdf-box-aws/internal/handler"

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

var _ handler.FileRepository = (*DynamoDbRepo)(nil)

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
		return nil, domain.ErrNotFound // no encontrar nada NO es error en Dynamo
	}

	var fi fileItem
	attributevalue.UnmarshalMap(out.Item, &fi)

	return fi.ToDomainFile(), nil
}

func (r *DynamoDbRepo) Save(ctx context.Context, file *domain.File) error {

	// TODO: implement PutItem, marshalling domain.File into DynamoDB attributes.
	return errors.New("not implemented")
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
		ScanIndexForward:       aws.Bool(false), // descendente
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
	}

	if cursor != "" {
		in.ExclusiveStartKey = repository.DecodeCursor(cursor)
	}

	out, err := r.client.Query(ctx, in)
	next := repository.EncodeCursor(out.LastEvaluatedKey) // "" si no hay más
	if err != nil {
		return nil, "", fmt.Errorf("dynamo query: %w", err)
	}

	files := make([]*domain.File, 0, len(out.Items))
	for _, item := range out.Items {
		var fi fileItem
		attributevalue.UnmarshalMap(item, &fi)
		files = append(files, fi.ToDomainFile())
	}

	return files, next, nil
}

func (r *DynamoDbRepo) MarkDeleted(ctx context.Context, userID, fileID string) error {
	// TODO: implement UpdateItem setting Status to domain.StatusDeleted.
	// Return domain.ErrNotFound when the item does not exist.
	return errors.New("not implemented")
}
