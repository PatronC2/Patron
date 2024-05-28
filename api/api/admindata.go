package api

import (
    "fmt"
    "net/http"
    "time"

    "golang.org/x/crypto/bcrypt"
    "github.com/gin-gonic/gin"
	"github.com/PatronC2/Patron/lib/logger"
    "github.com/PatronC2/Patron/types"
    "github.com/PatronC2/Patron/data"
    "github.com/jmoiron/sqlx"
)

var db *sqlx.DB

type User struct {
    types.User
}

func OpenDatabase(){ 
	var err error
	var port int
	host := data.GoDotEnvVariable("DB_HOST")
	fmt.Sscan(data.GoDotEnvVariable("DB_PORT"), &port)
	user := data.GoDotEnvVariable("DB_USER")
	password := data.GoDotEnvVariable("DB_PASS")
	dbname := data.GoDotEnvVariable("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
    "password=%s dbname=%s sslmode=disable",
    host, port, user, password, dbname)
	for {

		db, err = sqlx.Open("postgres", psqlInfo)
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
    defaultUserName := "patron"
    defaultUserPass := data.GoDotEnvVariable("ADMIN_AUTH_PASS")
    
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

func CreateUserHandler(c *gin.Context) {
    var req types.UserCreationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    user := &User{
        User: types.User{
            Username: req.Username,
            Role:     req.Role,
        },
    }

    if err := user.SetPassword(req.Password); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    logger.Logf(logger.Info, "Creating user in the database: %v", user.Username)
    if err := createUser(user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        logger.Logf(logger.Error, "Failed to create user: %v", err)
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "User created"})
    logger.Logf(logger.Info, "User %v created successfully", user.Username)
}

func DeleteUserByUsernameHandler(c *gin.Context) {
    username := c.Param("username")
    userID, err := GetUserIDByUsername(username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user ID"})
        return
    }
    err = DeleteUserByID(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
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
