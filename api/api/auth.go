package api

import (
	"errors"
	"net/http"
	"time"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/dgrijalva/jwt-go"
	"github.com/PatronC2/Patron/types"
)

var jwtKey []byte

func InitAuth() {
	jwtKeyStr := os.Getenv("JWT_KEY")
	jwtKey = []byte(jwtKeyStr)
}

func Auth(validRoles []string) gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := context.GetHeader("Authorization")
		if tokenString == "" {
			context.JSON(401, gin.H{"error": "request does not contain an access token"})
			context.Abort()
			return
		}
		err:= ValidateToken(tokenString, validRoles)
		if err != nil {
			context.JSON(401, gin.H{"error": err.Error()})
			context.Abort()
			return
		}
		context.Next()
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func GenerateJWT(username string, role string) (tokenString string, err error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims:= &types.JWTClaim{
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

func ValidateToken(signedToken string, validRoles []string) (err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&types.JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)

	if err != nil {
		return
	}

	claims, ok := token.Claims.(*types.JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return
	}

	if !contains(validRoles, claims.Role) {
		err = errors.New("Insufficient Privileges")
		return
	}	
	return

}

func ValidateAndGetClaims(tokenString string) (*types.JWTClaim, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&types.JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*types.JWTClaim)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func LoginHandler(c *gin.Context) {
    var loginRequest struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := c.ShouldBindJSON(&loginRequest); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    user, err := GetUserByUsername(loginRequest.Username)
    if err != nil || user.CheckPassword(loginRequest.Password) != nil {
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err})
		}
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    token, err := GenerateJWT(user.Username, user.Role)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": token})
}
