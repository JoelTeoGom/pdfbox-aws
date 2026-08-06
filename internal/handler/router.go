package handler

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type Router struct {
	API  *API
	auth TokenValidator
}

func NewRouter(api *API, auth TokenValidator) *Router {
	return &Router{
		API:  api,
		auth: auth,
	}
}
func (r *Router) handleRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	// Validate the JWT token from the request headers
	claims, err := r.auth.ValidateToken(req.Headers["Authorization"])
	if err != nil {
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: 401,
			Body:       "Unauthorized",
		}, nil
	}

	switch req.RouteKey {
	case "GET /files/{fileID}":
		return r.handleGetFile(ctx, claims.UserID, req)
	case "POST /files":
		return r.handleUploadFile(ctx, claims.UserID, req)
	case "GET /files":
		return r.handleListFiles(ctx, claims.UserID, req)
	case "DELETE /files/{fileID}":
		return r.handleDeleteFile(ctx, claims.UserID, req)
	default:
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: 404,
			Body:       "Not Found",
		}, nil
	}
}

func (r *Router) handleGetFile(ctx context.Context, userID string, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	fileID := req.PathParameters["fileID"]
	file, err := r.API.Get(ctx, userID, fileID)
	if err != nil {
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       "Internal Server Error",
		}, nil
	}
	if file == nil {
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: 404,
			Body:       "File Not Found",
		}, nil
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       file.Content, // Assuming file.Content is a string. Adjust as necessary.
	}, nil
}

func (r *Router) handleUploadFile(ctx context.Context, userID string, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	//TODO : Implement file upload logic
	return &events.APIGatewayV2HTTPResponse{
		StatusCode: 501,
		Body:       "Not Implemented",
	}, nil
}

func (r *Router) handleListFiles(ctx context.Context, userID string, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	//TODO : Implement file listing logic
	return &events.APIGatewayV2HTTPResponse{
		StatusCode: 501,
		Body:       "Not Implemented",
	}, nil
}

func (r *Router) handleDeleteFile(ctx context.Context, userID string, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	fileID := req.PathParameters["fileID"]
	err := r.API.MarkDeleted(ctx, userID, fileID)
	if err != nil {
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       "Internal Server Error",
		}, nil
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       "File Deleted",
	}, nil
}
