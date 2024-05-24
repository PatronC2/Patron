package data

import (
    "github.com/gin-gonic/gin"
    "github.com/dgrijalva/jwt-go"
    "net/http"
    "time"
)

var jwtSecret = []byte("your_secret_key")

func generateToken(username, role string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": username,
        "role":     role,
        "exp":      time.Now().Add(time.Hour * 72).Unix(),
    })
    return token.SignedString(jwtSecret)
}

func authMiddleware(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
            c.Abort()
            return
        }

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })

        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            if claims["role"] == role || claims["role"] == "admin" {
                c.Set("username", claims["username"])
                c.Set("role", claims["role"])
            } else {
                c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
                c.Abort()
                return
            }
        } else {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        c.Next()
    }
}
