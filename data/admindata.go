package data

import (
	"database/sql"
	"fmt"

	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
)

func UpdateUser(user *types.User) error {
	query := `
        UPDATE users 
        SET password_hash = COALESCE($1, password_hash), 
            role = COALESCE($2, role) 
        WHERE username = $3
    `
	_, err := db.Exec(query, user.PasswordHash, user.Role, user.Username)
	return err
}

func GetUserIDByUsername(username string) (int, error) {
	var userID int
	query := "SELECT id FROM users WHERE username = $1"
	err := db.QueryRow(query, username).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func DeleteUserByID(userID int) error {
	query := "DELETE FROM users WHERE id = $1"
	result, err := db.Exec(query, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		logger.Logf(logger.Error, "User could not be deleted")
	}
	return nil
}

func GetUserByUsername(username string) (*types.User, error) {
	var user types.User
	query := "SELECT id, username, password_hash, role FROM users WHERE username = $1"
	err := db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func UpdateUserPassword(user *types.User) error {
	query := "UPDATE users SET password_hash = $1 WHERE username = $2"
	_, err := db.Exec(query, user.PasswordHash, user.Username)
	return err
}

func CreateUser(user *types.User) error {
	CreateUserSQL := `
	INSERT INTO users (username, password_hash, role)
	VALUES ($1, $2, $3)
	ON CONFLICT (username) DO NOTHING;
	`

	result, err := db.Exec(CreateUserSQL, user.Username, user.PasswordHash, user.Role)
	if err != nil {
		logger.Logf(logger.Error, "Failed to create user: %v\n", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not determine if user was inserted: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user '%s' already exists", user.Username)
	}

	logger.Logf(logger.Info, "User %v created\n", user.Username)
	return nil
}

func UpsertUser(user *types.User) error {
	query := `
	INSERT INTO users (username, password_hash, role)
	VALUES ($1, $2, $3)
	ON CONFLICT (username) DO UPDATE
	SET password_hash = EXCLUDED.password_hash,
	    role = EXCLUDED.role;
	`
	_, err := db.Exec(query, user.Username, user.PasswordHash, user.Role)
	return err
}

func GetUsers() ([]types.User, error) {
	var users []types.User
	query := "SELECT id, username, role FROM users"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user types.User
		err := rows.Scan(&user.ID, &user.Username, &user.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
