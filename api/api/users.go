package api

import (
    "net/http"

    "golang.org/x/crypto/bcrypt"
    "github.com/gin-gonic/gin"
	"github.com/PatronC2/Patron/lib/logger"
    "github.com/PatronC2/Patron/types"
)

type User struct {
    types.User
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
    defaultUserPass := goDotEnvVariable("ADMIN_AUTH_PASS")
    
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
    // Extract username from request
    username := c.Param("username")

    // Retrieve user ID by username
    userID, err := GetUserIDByUsername(username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user ID"})
        return
    }

    // Delete user by ID
    err = DeleteUserByID(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
