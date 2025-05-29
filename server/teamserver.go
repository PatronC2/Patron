package main

import (
	"log"
	"time"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/server/quicListener"
	"github.com/PatronC2/Patron/server/tcpListener"
)

func Init() {
	enableLogging := true
	logger.EnableLogging(enableLogging)
	err := logger.SetLogFile("logs/server.log")
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

func main() {
	appName := "server"
	Init()
	data.OpenDatabase()
	data.InitDatabase()
	Refresh(appName)

	tlsServer := tcpListener.NewServer()
	go tlsServer.Start()

	quicServer := quicListener.NewServer()
	go func() {
		if err := quicServer.Start(); err != nil {
			logger.Logf(logger.Error, "QUIC server failed: %v", err)
		}
	}()

	select {}
}
