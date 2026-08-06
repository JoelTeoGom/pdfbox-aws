package api

import (
	"context"
	"os"
	"pdf-box-aws/internal/adapter/auth"
	"pdf-box-aws/internal/adapter/repository"
	"pdf-box-aws/internal/handler"
)

func main() {
	tableName, ok := os.LookupEnv("TABLE")
	secret, secretOk := os.LookupEnv("JWT_SECRET")
	if !ok {
		panic("Need TABLE environment variable")
	}
	if !secretOk {
		panic("Need JWT_SECRET environment variable")
	}
	ctx := context.Background()
	dynamodb := repository.NewDynamoDBStore(ctx, tableName)
	s3 := repository.NewS3Service()
	tokenValidator := auth.NewJWTService(secret)

	api := handler.NewAPI(dynamodb, tokenValidator, s3)
	router := handler.NewRouter(api, tokenValidator)
}
