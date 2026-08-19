package main

import (
	"context"
	"os"
	"pdf-box-aws/internal/adapter/repository"
	"pdf-box-aws/internal/adapter/storage"
	"pdf-box-aws/internal/handler"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	eventBusName, ok := os.LookupEnv("EVENT_BUS_NAME")
	if !ok {
		panic("Need EVENT_BUS_NAME environment variable")
	}

	tableName, ok := os.LookupEnv("TABLE")
	if !ok {
		panic("Need TABLE environment variable")
	}
	s3Bucket, ok := os.LookupEnv("S3_BUCKET")
	if !ok {
		panic("Need S3_BUCKET environment variable")
	}

	ctx := context.Background()
	dynamodb := repository.NewDynamoDBStore(ctx, tableName)
	s3 := storage.NewS3Service(ctx, s3Bucket)
	bus := handler.NewEventBridgeScheduler(ctx, eventBusName, s3, dynamodb)
	lambda.Start(bus.FileSweeperHandler)
}
