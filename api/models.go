package main

import (
    "fmt"

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
    fmt.Println("Plaintext password", password)
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
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

    logger.Logf(logger.Info, "Username %v\n", user.Username)
    logger.Logf(logger.Info, "User password hash %v\n", user.PasswordHash)
    logger.Logf(logger.Info, "User role %v\n", user.Role)
    _, err := db.Exec(CreateUserSQL, user.Username, user.PasswordHash, user.Role)
    if err != nil {
        logger.Logf(logger.Error, "Failed to create user: %v\n", err)
    }
    logger.Logf(logger.Info, "User %v created\n", user.Username)
	return err

}

func createAdminUser() error {
	default_user_name := "patron"
	default_user_pass := goDotEnvVariable("ADMIN_AUTH_PASS")
    user := &User{
        Username: default_user_name,
        Role:     "admin",
    }
    err := user.SetPassword(default_user_pass)
    if err != nil {
        return err
    }
    err = createUser(user)
    if err != nil {
        return err
    }
    logger.Logf(logger.Info, "User %v created\n", default_user_name)
    return err
}