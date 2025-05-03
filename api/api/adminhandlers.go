package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
	"github.com/gin-gonic/gin"
)

type UpdateUserRequest struct {
	types.UpdateUserRequest
}

func CreateAdminUser() error {
	defaultUserName := os.Getenv("ADMIN_AUTH_USER")
	defaultUserPass := os.Getenv("ADMIN_AUTH_PASS")

	if defaultUserName == "" || defaultUserPass == "" {
		return fmt.Errorf("ADMIN_AUTH_USER or ADMIN_AUTH_PASS not set")
	}

	user := &types.User{
		Username: defaultUserName,
		Role:     "admin",
	}

	if err := user.SetPassword(defaultUserPass); err != nil {
		return fmt.Errorf("failed to set password for admin user: %w", err)
	}

	err := data.UpsertUser(user)
	if err != nil {
		return fmt.Errorf("failed to create or update admin user: %w", err)
	}

	logger.Logf(logger.Info, "Admin user %v created or updated", defaultUserName)
	return nil
}

func DeleteUserByUsernameHandler(c *gin.Context) {
	defaultUserName := os.Getenv("ADMIN_AUTH_USER")
	username := c.Param("username")
	if defaultUserName == username {
		c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("user %s cannot be deleted", username)})
		return
	}
	userID, err := data.GetUserIDByUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user ID"})
		return
	}
	err = data.DeleteUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func CreateUserHandler(c *gin.Context) {
	var req types.UserCreationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user := &types.User{
		Username: req.Username,
		Role:     req.Role,
	}

	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	logger.Logf(logger.Info, "Creating user in the database: %v", user.Username)
	if err := data.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		logger.Logf(logger.Error, "Failed to create user: %v", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("Created user %s with role %s", user.Username, user.Role)})
	logger.Logf(logger.Info, "User %v created successfully", user.Username)
}

func UpdatePasswordHandler(c *gin.Context) {
	var passwordUpdateRequest struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}

	if err := c.ShouldBindJSON(&passwordUpdateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	tokenString := c.GetHeader("Authorization")
	claims, err := ValidateAndGetClaims(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	username := claims.Username
	user, err := data.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("User %s not found", username)})
		return
	}
	if err := user.CheckPassword(passwordUpdateRequest.OldPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid old password"})
		return
	}
	if err := user.SetPassword(passwordUpdateRequest.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}
	if err := data.UpdateUserPassword(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func UpdateUserHandler(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	defaultUserName := os.Getenv("ADMIN_AUTH_USER")
	username := c.Param("username")
	if defaultUserName == username {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("User %s cannot be modified", username)})
		return
	}

	user, err := data.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	if req.NewPassword != nil {
		if err := user.SetPassword(*req.NewPassword); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
	}

	if req.NewRole != nil && *req.NewRole != "" {
		user.Role = *req.NewRole
	}

	if err := data.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func GetUsersHandler(c *gin.Context) {
	users, err := data.GetUsers()
	if err != nil {
		logger.Logf(logger.Error, "Failed to get users: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func GetCurrentUserHandler(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	claims, err := ValidateAndGetClaims(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	username := claims.Username
	user, err := data.GetUserByUsername(username)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}
