package worker

import (
	"context"
	"os"
	"pdf-box-aws/internal/adapter/repository"
	"pdf-box-aws/internal/adapter/storage"
	"pdf-box-aws/internal/handler"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
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
	worker := handler.NewWorker(dynamodb, s3)
	lambda.Start(worker.HandleRequest)
}
