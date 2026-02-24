package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	errInvalidToken = errors.New("invalid token")
	errExpiredToken = errors.New("token expired")
)

type Claims struct {
	Sub      string `json:"sub"`
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}

type Manager struct {
	secret []byte
	ttl    time.Duration
}

func NewManager(secret string, ttl time.Duration) (*Manager, error) {
	trimmed := strings.TrimSpace(secret)
	if len(trimmed) < 16 {
		return nil, errors.New("auth secret must be at least 16 characters")
	}
	if ttl <= 0 {
		return nil, errors.New("token ttl must be positive")
	}
	return &Manager{secret: []byte(trimmed), ttl: ttl}, nil
}

func (m *Manager) IssueToken(userID, username string) (string, error) {
	if userID == "" || username == "" {
		return "", errors.New("user id and username required")
	}

	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	now := time.Now()
	claims := Claims{
		Sub:      userID,
		Username: username,
		Iat:      now.Unix(),
		Exp:      now.Add(m.ttl).Unix(),
	}

	headerPart, err := encodePart(header)
	if err != nil {
		return "", fmt.Errorf("encode header: %w", err)
	}
	claimsPart, err := encodePart(claims)
	if err != nil {
		return "", fmt.Errorf("encode claims: %w", err)
	}

	signingInput := headerPart + "." + claimsPart
	signature := signHS256(m.secret, signingInput)
	return signingInput + "." + signature, nil
}

func (m *Manager) ParseToken(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	expected := signHS256(m.secret, signingInput)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, errInvalidToken
	}

	payload, err := decodePart(parts[1])
	if err != nil {
		return nil, errInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, errInvalidToken
	}

	now := time.Now().Unix()
	if claims.Exp > 0 && now > claims.Exp {
		return nil, errExpiredToken
	}
	if claims.Iat > 0 && claims.Iat-now > 300 {
		return nil, errInvalidToken
	}

	return &claims, nil
}

func HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", errors.New("password required")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hashed), nil
}

func CheckPassword(hash, password string) bool {
	if hash == "" || password == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func encodePart(value any) (string, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func decodePart(part string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(part)
}

func signHS256(secret []byte, input string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
