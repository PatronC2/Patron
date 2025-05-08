package api

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
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
		err := ValidateToken(tokenString, validRoles)
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

func GenerateJWT(username string, role string, expirationDuration time.Duration) (tokenString string, err error) {
	// Default to 8 hours if no duration is provided
	if expirationDuration == 0 {
		expirationDuration = 8 * time.Hour
	}

	expirationTime := time.Now().Add(expirationDuration)
	claims := &types.JWTClaim{
		Username: username,
		Role:     role,
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
		logger.Logf(logger.Info, "Insufficient Prvileges")
		err = errors.New("insufficient privileges")
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
		Duration int64  `json:"duration"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := data.GetUserByUsername(loginRequest.Username)
	if err != nil || user.CheckPassword(loginRequest.Password) != nil {
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	var duration time.Duration
	if loginRequest.Duration > 0 {
		duration = time.Duration(loginRequest.Duration) * 24 * time.Hour
	} else {
		duration = 0
	}

	token, err := GenerateJWT(user.Username, user.Role, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
