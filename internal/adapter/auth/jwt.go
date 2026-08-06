package auth

import (
	"context"
	"crisplite/internal/domain"

	"fmt"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const claimsKey contextKey = "claims"

type JWTService struct {
	secretKey []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secretKey: []byte(secret)}
}

func (s *JWTService) GenerateAccessToken(claims *domain.Claims) (string, error) {
	JWTclaims := jwt.MapClaims{
		"user_id": claims.UserID,
		"role":    claims.Role,
		"exp":     jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		"iat":     jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTclaims)
	return token.SignedString(s.secretKey)
}

func (s *JWTService) ValidateToken(tokenString string) (*domain.Claims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected algorithm: %v", t.Header["alg"])
		}
		return s.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	userID, _ := mapClaims["user_id"].(string)
	role, _ := mapClaims["role"].(string)

	return &domain.Claims{UserID: userID, Role: role}, nil
}

func (s *JWTService) AddClaimsToContext(ctx context.Context, claims *domain.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func (s *JWTService) ClaimsFromContext(ctx context.Context) (*domain.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*domain.Claims)
	return claims, ok
}

func (s *JWTService) GenerateRefreshToken(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
