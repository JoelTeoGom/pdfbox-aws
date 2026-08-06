package domain

import "errors"

var (
	ErrNotFound      = errors.New("file not found")
	ErrNotReady      = errors.New("file not ready")
	ErrInvalidSize   = errors.New("invalid file size")
	ErrInvalidType   = errors.New("invalid content type")
	ErrInvalidStatus = errors.New("invalid status transition")
)
