package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService signs and verifies JWT tokens.
type JWTService struct {
	secretKey     []byte
	tokenLifetime time.Duration
}

// NewJWTService constructs a new JWT token service.
func NewJWTService(secretKey string, lifetime time.Duration) *JWTService {
	if lifetime <= 0 {
		lifetime = 24 * time.Hour
	}
	return &JWTService{
		secretKey:     []byte(secretKey),
		tokenLifetime: lifetime,
	}
}

// MintToken creates a signed JWT token string with the given claims.
func (s *JWTService) MintToken(userID int64, orgID string, roles []string) (string, error) {
	now := time.Now().UTC()
	claims := UserClaims{
		UserID: userID,
		OrgID:  orgID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenLifetime)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}
	return signed, nil
}

// ParseToken validates and extracts UserClaims from a JWT token string.
func (s *JWTService) ParseToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
