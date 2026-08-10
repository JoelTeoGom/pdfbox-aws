package worker

import (
	"context"
	"os"
	"pdf-box-aws/internal/adapter/repository"
	"pdf-box-aws/internal/handler"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	tableName, ok := os.LookupEnv("TABLE")
	if !ok {
		panic("Need TABLE environment variable")
	}

	ctx := context.Background()
	dynamodb := repository.NewDynamoDBStore(ctx, tableName)
	worker := handler.NewWorker(dynamodb)
	lambda.Start(worker.HandleRequest)
}
