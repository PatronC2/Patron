package api

import (
    "database/sql"
    "fmt"
    "time"
    "os"

    "golang.org/x/crypto/bcrypt"
	"github.com/PatronC2/Patron/lib/logger"
    "github.com/PatronC2/Patron/types"
)

var db *sql.DB

type User struct {
    types.User
}

func OpenDatabase(){ 
	var err error
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASS")
    dbname := os.Getenv("DB_NAME")

    logger.Logf(logger.Info, "Got environment variables host=%s, port=%s, user=%s, dbname=%s (password not shown)", host, port, user, dbname)

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
    "password=%s dbname=%s sslmode=disable",
    host, port, user, password, dbname)
	for {

		db, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			logger.Logf(logger.Error, "Failed to connect to the database: %v\n", err)
			time.Sleep(30 * time.Second)
			continue
		}
		err = db.Ping()
		if err != nil {
			logger.Logf(logger.Error, "Failed to ping the database: %v\n", err)
			db.Close()
			time.Sleep(30 * time.Second)
			continue
		}
		logger.Logf(logger.Info, "Postgres DB connected\n")
		break
	}
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

func CreateAdminUser() error {
    defaultUserName := os.Getenv("ADMIN_AUTH_USER")
    defaultUserPass := os.Getenv("ADMIN_AUTH_PASS")
    
    user := &User{
        User: types.User{
            Username: defaultUserName,
            Role:     "admin",
        },
    }
    
    err := user.SetPassword(defaultUserPass)
    if err != nil {
        return err
    }
    
    err = createUser(user)
    if err != nil {
        return err
    }
    
    logger.Logf(logger.Info, "User %v created\n", defaultUserName)
    return nil
}

func updateUser(user *User) error {
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

func GetUserByUsername(username string) (*User, error) {
    var user User
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


func updateUserPassword(user *User) error {
	query := "UPDATE users SET password_hash = $1 WHERE username = $2"
	_, err := db.Exec(query, user.PasswordHash, user.Username)
	return err
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

func GetUsers() ([]User, error) {
    var users []User
    query := "SELECT id, username, role FROM users"
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var user User
        err := rows.Scan(&user.ID, &user.Username, &user.Role)
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }
    return users, nil
}

