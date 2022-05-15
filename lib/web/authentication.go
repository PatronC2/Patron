package web

import (
	"database/sql"
	"errors"

	//"html/template"

	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt"

	"github.com/PatronC2/Patron/lib/logger"
)

type UserPermissions string

const (
	PERMS_USER  UserPermissions = "user"
	PERMS_ADMIN UserPermissions = "admin"
)

// TODO: Read jwtSigningKey from a config file (.env) that won't be pushed to Git
var (
	jwtSigningKey    []byte = []byte("supersecretsecret")
	CookieExpiration int64  = time.Now().Add(time.Minute * 30).Unix() // token expires after 30 minutes
)

/*
	Parses the JSON Web Token "auth" cookie and returns its claims as
	jwt.MapClaims.
*/
func GetAuthClaims(writer http.ResponseWriter, request *http.Request) (authClaims jwt.MapClaims, err error) {
	// Get auth cookie
	authCookie, err := request.Cookie("auth")
	// If cookie doesn't exist
	if err != nil {
		http.Redirect(writer, request, "/login", http.StatusFound)
		return
	}

	authClaims, err = GetJWTClaims(authCookie, writer, request)
	return
}

/*
	Parses the provided JSON Web Token cookie and returns its claims as
	jwt.MapClaims.
*/
func GetJWTClaims(jwtCookie *http.Cookie, writer http.ResponseWriter, request *http.Request) (tokenClaims jwt.MapClaims, err error) {
	// Parse token
	token, err := jwt.Parse(
		jwtCookie.Value,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("incorrect signing method")
			}

			// Return the signing key for token validation
			return jwtSigningKey, nil
		},
	)

	// Token is not a proper JWT, is expired, or does not use the correct
	// signing method
	if err != nil {
		return
	}

	// I believe the above jwt.Parse() already does all of the token validation
	// for us, but the below checks are "just in case".

	// Check if the JWT is valid
	if !token.Valid {
		err = errors.New("token invalid")
		return
	}

	// Type the claims as jwt.MapClaims
	if mapClaims, ok := token.Claims.(jwt.MapClaims); !ok {
		err = errors.New("token claims invalid")
		return
	} else {
		// Make sure the token isn't expired (current time > exp).
		// Automatically uses the RFC standard "exp", "iat", and "nbf" claims,
		// if they exist, with values interpretted as UNIX seconds.
		err = mapClaims.Valid()
		if err != nil {
			return
		}

		tokenClaims = token.Claims.(jwt.MapClaims)
	}

	if len(tokenClaims) == 0 {
		err = errors.New("token claims don't exist")
		return
	}

	return
}

func GenerateJWT(db *sql.DB, username, perms UserPermissions) (tokenStirng string, err error) {
	/*
		--- Token Payload (Claims) ---
		"user": <team_name>,
		"perms": <user_rank>,
		"exp": <timestamp_unix_seconds>

		JWT is stored as a cookie with the name "auth"
	*/

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["user"] = username
	claims["perms"] = perms
	claims["exp"] = CookieExpiration

	tokenString, err := token.SignedString(jwtSigningKey)

	return tokenString, err
}

func ValidatePasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil // err == nil means incorrect password
}

func HashPassword(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		logger.Log(logger.Error, "Could not hash provided password")
		logger.LogError(err)
	}

	return string(hashBytes), err
}
