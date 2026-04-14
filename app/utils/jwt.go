package utils

import (
	"context"
	"missfit/app/facades"
	"missfit/app/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte(facades.Config().GetString("jwt.secret"))

type contextKey string

const userKey = contextKey("user")

func GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ParseToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", err
	}

	return userID, nil
}

func GetUser(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userKey).(*models.User)
	return user, ok
}
