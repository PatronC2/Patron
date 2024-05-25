package main

import (
    "golang.org/x/crypto/bcrypt"
	"github.com/PatronC2/Patron/lib/logger"	
)

type User struct {
    ID           int    `db:"id"`
    Username     string `db:"username"`
    PasswordHash string `db:"password_hash"`
    Role         string `db:"role"`
}

func (u *User) SetPassword(password string) error {
    hash, err := HashPassword(password)
    if err != nil {
        logger.Logf(logger.Info, "Error hashing password: %v\n", err)
    }
    u.PasswordHash = string(hash)
    return nil
}

func (u *User) CheckPassword(password string) error {
    return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

func getUserByUsername(username string) (*User, error) {
    var user User
    err := db.Get(&user, "SELECT * FROM users WHERE username=$1", username)
    return &user, err
}

func createUser(user *User) error {
    CreateUserSQL := `
	INSERT INTO users (username, password_hash, role)
	VALUES ($1, $2, $3)
	ON CONFLICT (username) DO NOTHING;
	`

    _, err := db.Exec(CreateUserSQL, user.Username, user.PasswordHash, user.Role)
    if err != nil {
        logger.Logf(logger.Error, "Failed to create user: %v\n", err)
    }
    logger.Logf(logger.Info, "User %v created\n", user.Username)
	return err

}
