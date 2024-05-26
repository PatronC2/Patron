package main

import (
    "fmt"

    "github.com/PatronC2/Patron/lib/logger"
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    OpenDatabase()
    InitDatabase()
    createAdminUser()
    
    r := gin.Default()

    r.POST("/login", loginHandler)
    r.POST("/users", Auth("admin"), createUserHandler)
    r.DELETE("/users/:username", Auth("admin"), deleteUserByUsernameHandler)

    api := r.Group("/api")
    api.Use(Auth("readOnly"))
    {
        api.GET("/data", readOnlyHandler)
    }

    admin := r.Group("/admin")
    admin.Use(Auth("admin"))
    {
        admin.POST("/data", adminHandler)
    }

    // Replace with your paths to the certificate and key files
    certFile := "certs/server.pem"
    keyFile := "certs/server.key"

    // Use RunTLS to enable SSL
    if err := r.RunTLS(":8443", certFile, keyFile); err != nil {
        logger.Logf(logger.Error, "Failed to run server: %v\n", err)
    }
}

func loginHandler(c *gin.Context) {
    var loginRequest struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := c.ShouldBindJSON(&loginRequest); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    user, err := getUserByUsername(loginRequest.Username)
	fmt.Println("loginHandler getUser", user)
    if err != nil || user.CheckPassword(loginRequest.Password) != nil {
		if err != nil {
			fmt.Println("login error", err)
		}
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    token, err := GenerateJWT(user.Username, user.Role)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": token})
}

func readOnlyHandler(c *gin.Context) {
    // Implement read-only access logic here
    c.JSON(http.StatusOK, gin.H{"data": "read-only data"})
}

func adminHandler(c *gin.Context) {
    // Implement admin read/write access logic here
	c.JSON(http.StatusOK, gin.H{"data": "admmin data"})
}
