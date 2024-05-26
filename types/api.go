package types

import (
	"github.com/dgrijalva/jwt-go"
)
type JWTClaim struct {
	Username string `json:"username"`
	Role    string `json:"role"`
	jwt.StandardClaims
}

type User struct {
    ID           int    `db:"id"`
    Username     string `db:"username"`
    PasswordHash string `db:"password_hash"`
    Role         string `db:"role"`
}

type UserCreationRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
    Role     string `json:"role" binding:"required"`
}
