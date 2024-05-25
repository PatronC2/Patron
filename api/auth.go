package main

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"github.com/PatronC2/Patron/lib/logger"
)

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


func HashPassword(plaintextPassword string) (passwordHash string, err error) {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
    if err != nil {
        logger.Logf(logger.Info, "Error hashing password: %v\n", err)
        return "", err
    }
    return string(hashedBytes), nil
}
