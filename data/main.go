package data

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    initDB()
    r := gin.Default()

    r.POST("/login", loginHandler)
    r.POST("/users", authMiddleware("admin"), createUserHandler)

    api := r.Group("/api")
    api.Use(authMiddleware("read-only"))
    {
        api.GET("/data", readOnlyHandler)
    }

    admin := r.Group("/admin")
    admin.Use(authMiddleware("admin"))
    {
        admin.POST("/data", adminHandler)
    }

    r.Run(":8080")
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
    if err != nil || user.CheckPassword(loginRequest.Password) != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    token, err := generateToken(user.Username, user.Role)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": token})
}

func createUserHandler(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    if err := user.SetPassword(user.PasswordHash); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
        return
    }

    if err := createUser(&user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "User created"})
}

func readOnlyHandler(c *gin.Context) {
    // Implement read-only access logic here
    c.JSON(http.StatusOK, gin.H{"data": "read-only data"})
}

func operatorHandler(c *gin.Context) {
    // Implement operator read/write access logic here
    c.JSON(http.StatusOK, gin.H{"data": "operator data"})
}

func adminHandler(c *gin.Context) {
    // Implement admin read/write access logic here
    c.JSON(http.StatusOK, gin.H{"data": "admin data"})
}
