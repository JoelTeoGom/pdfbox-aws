package api

import (
	"context"
	"os"
	"pdf-box-aws/internal/adapter/auth"
	"pdf-box-aws/internal/adapter/repository"
	"pdf-box-aws/internal/adapter/storage"
	"pdf-box-aws/internal/handler"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	tableName, ok := os.LookupEnv("TABLE")
	secret, secretOk := os.LookupEnv("JWT_SECRET")
	bucket, bucketOk := os.LookupEnv("S3_BUCKET")
	if !ok {
		panic("Need TABLE environment variable")
	}
	if !secretOk {
		panic("Need JWT_SECRET environment variable")
	}
	if !bucketOk {
		panic("Need S3_BUCKET environment variable")
	}

	ctx := context.Background()
	dynamodb := repository.NewDynamoDBStore(ctx, tableName)
	s3 := storage.NewS3Service(ctx, bucket)
	tokenValidator := auth.NewJWTService(secret)

	api := handler.NewAPI(dynamodb, s3)
	router := handler.NewRouter(api, tokenValidator)
	lambda.Start(router.HandleRequest)
}
