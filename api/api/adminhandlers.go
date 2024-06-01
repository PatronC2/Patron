package api

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
	"github.com/PatronC2/Patron/lib/logger"
    "github.com/PatronC2/Patron/types"
	"github.com/PatronC2/Patron/data"
)

type UpdateUserRequest struct {
    types.UpdateUserRequest
}

func DeleteUserByUsernameHandler(c *gin.Context) {
	defaultUserName := data.GoDotEnvVariable("ADMIN_AUTH_USER")
    username := c.Param("username")
	if defaultUserName == username {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("user %s cannot be deleted", username)})
		return
	}
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

    c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("Created user %s with role %s", user.Username, user.Role),})
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
	user, err := GetUserByUsername(username)
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
	if err := updateUserPassword(user); err != nil {
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

    defaultUserName := data.GoDotEnvVariable("ADMIN_AUTH_USER")
    username := c.Param("username")
    if defaultUserName == username {
        c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("User %s cannot be modified", username)})
        return
    }

    user, err := GetUserByUsername(username)
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

    if err := updateUser(user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func GetUsersHandler(c *gin.Context) {
    users, err := GetUsers()
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
    user, err := GetUserByUsername(username)
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
