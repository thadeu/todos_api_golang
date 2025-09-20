package auth

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	Secret string
}

func (j *JWT) CreateToken(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(time.Hour * 3).Unix(),
	})

	return token.SignedString([]byte(j.Secret))
}

func (j *JWT) VerifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(j.Secret), nil
	})

	if err != nil {
		slog.Error("Error verifying token", "error", err)
		return nil, err
	}

	if !token.Valid {
		slog.Error("Invalid access token")
		return nil, fmt.Errorf("%v", "Invalid Access Token")
	}

	claims := token.Claims.(jwt.MapClaims)

	return claims, nil
}

func CreateJwtTokenForUser(userId int) (string, error) {
	jwt := JWT{Secret: os.Getenv("JWT_SECRET")}
	return jwt.CreateToken(userId)
}

func VerifyJwtToken(token string) (jwt.MapClaims, error) {
	jwt := JWT{Secret: os.Getenv("JWT_SECRET")}
	return jwt.VerifyToken(token)
}
