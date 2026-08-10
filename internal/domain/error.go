package domain

import "errors"

var (
	// Input validation
	ErrInvalidInput    = errors.New("invalid input")
	ErrFileTooLarge    = errors.New("file exceeds the maximum allowed size")
	ErrInvalidMimeType = errors.New("invalid content type")

	// Lookup and access
	ErrFileNotFound      = errors.New("file not found")
	ErrFileAlreadyExists = errors.New("file already exists")
	ErrUnauthorized      = errors.New("unauthorized")

	// Lifecycle
	ErrNotReady         = errors.New("file not ready")
	ErrInvalidStatus    = errors.New("invalid status transition")
	ErrAlreadyProcessed = errors.New("file already processed")
)
