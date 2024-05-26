package main

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/dgrijalva/jwt-go"
	"github.com/PatronC2/Patron/types"
)

// TODO: Pull this secret from .env
var jwtKey = []byte("patronsecret")

func Auth(requiredRole string) gin.HandlerFunc{
	return func(context *gin.Context) {
		tokenString := context.GetHeader("Authorization")
		if tokenString == "" {
			context.JSON(401, gin.H{"error": "request does not contain an access token"})
			context.Abort()
			return
		}
		err:= ValidateToken(tokenString, requiredRole)
		if err != nil {
			context.JSON(401, gin.H{"error": err.Error()})
			context.Abort()
			return
		}
		context.Next()
	}
}

func GenerateJWT(username string, role string) (tokenString string, err error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims:= &JWTClaim{
		Username: username,
		Role: role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(jwtKey)
	return
}

func ValidateToken(signedToken string, role string) (err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)

	if err != nil {
		return
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return
	}

	if claims.Role != role && claims.Role != "admin" {
		err = errors.New("Insufficient Privileges")
		return
	}
	
	return

}