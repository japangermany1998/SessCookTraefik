package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

func handle_when_invalid_token(c *fiber.Ctx, e error) error {
	fmt.Println("Error when parsing JWT", e)
	return e
}

func parseTokenClaims(tokenString string) (jwt.MapClaims, error){
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{},func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})

	if err != nil || !token.Valid{
		return nil, err
	}

	claims := token.Claims.(jwt.MapClaims)
	return claims, nil
}
