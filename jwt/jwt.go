package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateAccessToken(login string) (token string, err error) {
	t := jwt.New(jwt.SigningMethodHS256)
	mapClaims := t.Claims.(jwt.MapClaims)
	mapClaims["login"] = login
	mapClaims["exp"] = time.Now().Add(time.Minute * 30).Unix()
	token, err = t.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}

	return token, err
}

func ParseAccessToken(token string) (jwt.MapClaims, error) {
	claims, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error in parsing")
		}
		return []byte("secret"), nil
	})
	if err != nil {
		return nil, err
	}

	return claims.Claims.(jwt.MapClaims), err
}

func GenerateRefreshToken(id string) (token string, err error) {
	t := jwt.New(jwt.SigningMethodHS256)
	mapClaims := t.Claims.(jwt.MapClaims)
	mapClaims["id"] = id
	mapClaims["exp"] = time.Now().Add(time.Minute * 30).Unix()
	token, err = t.SignedString([]byte("refresh_token"))
	if err != nil {
		return "", err
	}

	return token, err
}

func ParseRefreshTokenToken(token string) (jwt.MapClaims, error) {
	claims, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error in parsing")
		}
		return []byte("refresh_token"), nil
	})
	if err != nil {
		return nil, err
	}

	return claims.Claims.(jwt.MapClaims), err
}
