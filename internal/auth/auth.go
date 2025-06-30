package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 1)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPasswordHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   fmt.Sprintf("%v", userID),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (userID uuid.UUID, err error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil })
	if err != nil {
		return uuid.UUID{}, err
	}

	idString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}
	userID, err = uuid.Parse(idString)
	if err != nil {
		return uuid.UUID{}, err
	}
	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers["Authorization"]
	if len(authHeader) == 0 {
		return "", fmt.Errorf("malformed authentication token")
	}
	prefix, stringToken, _ := strings.Cut(authHeader[0], " ")
	if prefix != "Bearer" {
		return "", fmt.Errorf("invalid authentication token: %v", authHeader[0])
	}

	return stringToken, nil
}

func MakeRefreshToken() (string, error) {
	rawRefreshToken := make([]byte, 32)
	rand.Read(rawRefreshToken)
	return hex.EncodeToString(rawRefreshToken), nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers["Authorization"]
	if len(authHeader) == 0 {
		return "", fmt.Errorf("malformed authentication key")
	}

	prefix, APIKey, _ := strings.Cut(authHeader[0], " ")
	if prefix != "ApiKey" {
		return "", fmt.Errorf("invalid authentication key: %v", authHeader[0])
	}

	return APIKey, nil
}
