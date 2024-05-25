package main

import (
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    ID           int    `db:"id"`
    Username     string `db:"username"`
    PasswordHash string `db:"password_hash"`
    Role         string `db:"role"`
}

func (u *User) SetPassword(password string) error {
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
    _, err := db.NamedExec(`INSERT INTO users (username, password_hash, role) 
                             VALUES (:username, :password_hash, :role)`, user)
    return err
}
