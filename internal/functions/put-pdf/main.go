package main

import (
	"context"
	"os"
	"pdf-box-aws/internal/adapter/repository"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	tableName, ok := os.LookupEnv("TABLE")
	if !ok {
		panic("Need TABLE environment variable")
	}

	ctx := context.Background()
	dynamodb := repository.NewDynamoDBStore(ctx, tableName)

	lambda.Start(handler.DeleteHandler)
}
