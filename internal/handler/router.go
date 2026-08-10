package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100

	bearerPrefix = "Bearer "
)

// bearerToken pulls the raw JWT out of the Authorization header. API Gateway
// HTTP APIs deliver header names lowercased, but the canonical spelling is
// accepted too so the handler also works behind a REST API or in tests.
func bearerToken(headers map[string]string) string {
	authorizationHeader := headers["authorization"]
	if authorizationHeader == "" {
		authorizationHeader = headers["Authorization"]
	}

	if len(authorizationHeader) >= len(bearerPrefix) &&
		strings.EqualFold(authorizationHeader[:len(bearerPrefix)], bearerPrefix) {
		authorizationHeader = authorizationHeader[len(bearerPrefix):]
	}

	return strings.TrimSpace(authorizationHeader)
}

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
func (r *Router) HandleRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	// Validate the JWT token from the request headers
	claims, err := r.auth.ValidateToken(bearerToken(req.Headers))
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
	file, presignedURL, err := r.API.Get(ctx, userID, fileID)
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

	body, err := json.Marshal(FileResponse{
		PresignedURL: presignedURL,
		File:         toFileDTO(file),
	})
	if err != nil {
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       "Internal Server Error",
		}, nil
	}
	return &events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}

func (r *Router) handleUploadFile(ctx context.Context, userID string, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	var uploadReq struct {
		Filename string `json:"filename"`
		Size     int64  `json:"size"`
		Mime     string `json:"mime"`
	}
	if err := json.Unmarshal([]byte(req.Body), &uploadReq); err != nil {
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       "Invalid Request Body",
		}, nil
	}

	file, presignedURL, err := r.API.Save(ctx, userID, uploadReq.Filename, uploadReq.Size, uploadReq.Mime)
	if err != nil {
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       "Internal Server Error",
		}, nil
	}

	body, err := json.Marshal(FileResponse{
		PresignedURL: presignedURL,
		File:         toFileDTO(file),
	})
	if err != nil {
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       "Internal Server Error",
		}, nil
	}
	return &events.APIGatewayV2HTTPResponse{
		StatusCode: 201,
		Body:       string(body),
	}, nil
}

func (r *Router) handleListFiles(ctx context.Context, userID string, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	limit := defaultListLimit
	if raw := req.QueryStringParameters["limit"]; raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return &events.APIGatewayV2HTTPResponse{
				StatusCode: 400,
				Body:       "Invalid limit",
			}, nil
		}
		limit = min(parsed, maxListLimit)
	}

	files, nextCursor, err := r.API.List(ctx, userID, req.QueryStringParameters["cursor"], limit)
	if err != nil {
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       "Internal Server Error",
		}, nil
	}

	body, err := json.Marshal(ListResponse{
		Items:      toFileDTOs(files),
		NextCursor: nextCursor,
	})
	if err != nil {
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       "Internal Server Error",
		}, nil
	}
	return &events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       string(body),
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
