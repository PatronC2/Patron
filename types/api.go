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
