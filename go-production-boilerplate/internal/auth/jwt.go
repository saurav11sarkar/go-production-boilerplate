package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/domain"
)

type Claims struct {
	Role         domain.Role `json:"role"`
	TokenVersion int         `json:"tv"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	issuer string
	secret []byte
	ttl    time.Duration
}

func NewJWTManager(issuer, secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{issuer: issuer, secret: []byte(secret), ttl: ttl}
}

func (m *JWTManager) CreateAccessToken(userID uuid.UUID, role domain.Role, tokenVersion int) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.ttl)
	claims := Claims{
		Role:         role,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	return signed, expiresAt, err
}

func (m *JWTManager) ParseAccessToken(raw string) (uuid.UUID, domain.Role, int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		raw,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
			}
			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil || !token.Valid {
		return uuid.Nil, "", 0, fmt.Errorf("invalid access token: %w", err)
	}
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, "", 0, fmt.Errorf("invalid subject: %w", err)
	}
	if !claims.Role.Valid() {
		return uuid.Nil, "", 0, fmt.Errorf("invalid role")
	}
	if claims.TokenVersion < 0 {
		return uuid.Nil, "", 0, fmt.Errorf("invalid token version")
	}
	return id, claims.Role, claims.TokenVersion, nil
}
