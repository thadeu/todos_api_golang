package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

func JwtAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")

		if bearer == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{"errors": []string{"Unauthorized request"}})
			return
		}

		token, err := VerifyJwtToken(bearer[len("Bearer "):])

		if err != nil {
			slog.Info("Error", "error", err)

			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{"errors": []string{"Unauthorized request", err.Error()}})
			return
		}

		userId := int(token["user_id"].(float64))
		context := context.WithValue(r.Context(), "x-user-id", userId)

		next.ServeHTTP(w, r.WithContext(context))
	}
}

func GinJwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearer := c.GetHeader("Authorization")

		if bearer == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errors": []string{"Unauthorized request"},
			})

			c.Abort()
			return
		}

		if !strings.HasPrefix(bearer, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errors": []string{"Invalid authorization format"},
			})

			c.Abort()
			return
		}

		token, err := VerifyJwtToken(bearer[len("Bearer "):])

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errors": []string{"Unauthorized request", err.Error()},
			})
			c.Abort()
			return
		}

		userId := int(token["user_id"].(float64))

		c.Set("x-user-id", userId)
		c.Next()
	}
}
