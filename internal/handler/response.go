package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"pdf-box-aws/internal/domain"

	"github.com/aws/aws-lambda-go/events"
)

// ErrorResponse is the body every failed request returns.
type ErrorResponse struct {
	Error string `json:"error"`
}

// statusFor maps a domain error to the status code the client should see. Any
// error that is not part of the domain is a fault on our side, so it falls
// through to 500 rather than being guessed at.
func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrFileNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrInvalidInput), errors.Is(err, domain.ErrInvalidMimeType):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrFileTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, domain.ErrFileAlreadyExists),
		errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrAlreadyProcessed):
		return http.StatusConflict
	case errors.Is(err, domain.ErrNotReady):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// jsonResponse serialises a payload with the JSON content type attached. API
// Gateway serves a response without it as text/plain.
func jsonResponse(statusCode int, payload any) *events.APIGatewayV2HTTPResponse {
	body, err := json.Marshal(payload)
	if err != nil {
		// Marshalling our own DTOs should not fail. If it ever does, answer with
		// valid JSON anyway rather than an empty body the client cannot parse.
		return &events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       `{"error":"internal server error"}`,
		}
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}
}

// errorResponse turns an error into the response the client sees. Domain errors
// carry a message that is safe to expose; anything else could leak table names,
// key layouts or SDK internals, so it is logged and reported as a bare 500.
func errorResponse(ctx context.Context, err error) *events.APIGatewayV2HTTPResponse {
	statusCode := statusFor(err)

	message := err.Error()
	if statusCode == http.StatusInternalServerError {
		slog.ErrorContext(ctx, "request failed", "error", err)
		message = "internal server error"
	}

	return jsonResponse(statusCode, ErrorResponse{Error: message})
}

// badRequest reports a malformed request, which never originates from a domain
// error because the request fails before reaching the use cases.
func badRequest(message string) *events.APIGatewayV2HTTPResponse {
	return jsonResponse(http.StatusBadRequest, ErrorResponse{Error: message})
}
