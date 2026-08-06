package domain

import "time"

type Claims struct {
	UserID string
	Role   string
}

type RefreshToken struct {
	ID          string
	HashedToken string
	UserID      string
	Revoked     bool
	CreatedAt   time.Time
}
