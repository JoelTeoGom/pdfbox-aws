package repository

import (
	"context"
	"errors"
	"log"
	"pdf-box-aws/internal/domain"
	"pdf-box-aws/internal/handler"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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
	// TODO: implement GetItem against the DynamoDB table and map the item to domain.File.
	// Return domain.ErrNotFound when the item does not exist.
	return nil, errors.New("not implemented")
}

func (r *DynamoDbRepo) Save(ctx context.Context, file *domain.File) error {
	// TODO: implement PutItem, marshalling domain.File into DynamoDB attributes.
	return errors.New("not implemented")
}

func (r *DynamoDbRepo) List(ctx context.Context, userID, cursor string, limit int) ([]*domain.File, string, error) {
	// TODO: implement Query by userID with pagination, decoding cursor into ExclusiveStartKey
	// and encoding LastEvaluatedKey into the returned cursor.
	return nil, "", errors.New("not implemented")
}

func (r *DynamoDbRepo) MarkDeleted(ctx context.Context, userID, fileID string) error {
	// TODO: implement UpdateItem setting Status to domain.StatusDeleted.
	// Return domain.ErrNotFound when the item does not exist.
	return errors.New("not implemented")
}
