package handler

import (
	"context"
	"log"
	"pdf-box-aws/internal/adapter/repository"
	"pdf-box-aws/internal/adapter/storage"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchevents"
)

type EventBridgeScheduler struct {
	client  *cloudwatchevents.Client
	busName string
}

func NewEventBridgeScheduler(ctx context.Context, busName string, s3 *storage.S3Service, dynamo *repository.DynamoDbRepo) *EventBridgeScheduler {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	client := cloudwatchevents.NewFromConfig(cfg)

	return &EventBridgeScheduler{
		client:  client,
		busName: busName,
	}
}

func (e *EventBridgeScheduler) FileSweeperHandler(ctx context.Context) error {
}
