// Utility functions for logging messages and errors
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type logType int

const (
	Error   logType = 0
	Warning logType = 1
	Info    logType = 2
	List    logType = 3
	Done    logType = 4
	Debug   logType = 5

	SUCCESS     int = 0
	ERR_GENERIC int = 1
	ERR_UNKNOWN int = 2

	ERR_USAGE     int = 10
	ERR_INPUT     int = 11
	ERR_FILE_READ int = 12
	ERR_CLOSABLE  int = 13

	ERR_CONNECTION int = 30
	ERR_WRITE      int = 31
	ERR_PARSE      int = 32
)

var MapTypesToColor = map[logType]*color.Color{
	Error:   color.New(color.Bold, color.FgRed),
	Warning: color.New(color.Bold, color.FgYellow),
	Info:    color.New(color.Bold, color.FgCyan),
	List:    color.New(color.Bold, color.FgBlue),
	Done:    color.New(color.Bold, color.FgGreen),
	Debug:   color.New(color.Bold, color.FgMagenta),
}

var MapTypesToPrefix = map[logType]string{
	Error:   MapTypesToColor[Error].Sprint("[ERR]"),
	Warning: MapTypesToColor[Warning].Sprint("[WRN]"),
	Info:    MapTypesToColor[Info].Sprint("[INF]"),
	List:    MapTypesToColor[List].Sprint("[LST]"),
	Done:    MapTypesToColor[Done].Sprint("[DON]"),
	Debug:   MapTypesToColor[Debug].Sprint("[DBG]"),
}

var (
	enabled                 = true
	currentLogLevel logType = Info
	logFile         *os.File
	logger          *log.Logger
	mu              sync.Mutex
)

// EnableLogging enables or disables logging globally.
func EnableLogging(flag bool) {
	mu.Lock()
	defer mu.Unlock()
	enabled = flag
}

// SetLogLevel sets the current minimum log level to show
func SetLogLevel(level logType) {
	mu.Lock()
	defer mu.Unlock()
	currentLogLevel = level
}

// SetLogLevelByName sets the log level using a string (e.g., "info", "error").
func SetLogLevelByName(level string) {
	switch strings.ToLower(level) {
	case "debug":
		SetLogLevel(Debug)
	case "info":
		SetLogLevel(Info)
	case "warning":
		SetLogLevel(Warning)
	case "error":
		SetLogLevel(Error)
	default:
		SetLogLevel(Info)
	}
}

// SetLogFile sets the file to which logs will be written.
func SetLogFile(filename string) error {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		logFile.Close()
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	logFile = file
	logger = log.New(io.MultiWriter(os.Stdout, file), "", log.LstdFlags)
	return nil
}

func shouldLog(level logType) bool {
	return level <= currentLogLevel
}

// Log a timestamped message with a given logType.
func Log(messageType logType, messages ...string) {
	mu.Lock()
	defer mu.Unlock()

	if enabled && shouldLog(messageType) {
		fmt.Printf("%s (%s) - %s\n",
			MapTypesToPrefix[messageType],
			time.Now().Format(time.RFC3339),
			MapTypesToColor[messageType].Sprint(strings.Join(messages, " ")),
		)
	}
}

// Logf functions like fmt.Printf for a given logType.
func Logf(messageType logType, format string, data ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	if enabled && shouldLog(messageType) {
		logString := fmt.Sprintf("%s (%s) - %s",
			MapTypesToPrefix[messageType],
			time.Now().Format(time.RFC3339),
			MapTypesToColor[messageType].Sprintf(format, data...),
		)
		if logger != nil {
			logger.Print(logString)
		}
	}
}

// LogPlain logs a message without a timestamp.
func LogPlain(messageType logType, messages ...string) {
	mu.Lock()
	defer mu.Unlock()

	if enabled && shouldLog(messageType) {
		fmt.Printf("%s %s\n", MapTypesToPrefix[messageType], MapTypesToColor[messageType].Sprint(strings.Join(messages, " ")))
	}
}

// LogReturn returns the formatted log string instead of printing it.
func LogReturn(messageType logType, messages ...string) string {
	mu.Lock()
	defer mu.Unlock()

	if enabled && shouldLog(messageType) {
		return fmt.Sprintf("%s (%s) - %s",
			MapTypesToPrefix[messageType],
			time.Now().Format(time.RFC3339),
			MapTypesToColor[messageType].Sprint(strings.Join(messages, " ")),
		)
	} else {
		return ""
	}
}

// LogError logs the provided error as an error message.
func LogError(err error) {
	Log(Error, err.Error())
}

func TruncateLogFileIfTooLarge(maxBytes int64) error {
	mu.Lock()
	defer mu.Unlock()

	if logFile == nil {
		return fmt.Errorf("log file not initialized")
	}

	info, err := logFile.Stat()
	if err != nil {
		return err
	}

	if info.Size() <= maxBytes {
		return nil
	}

	offset := info.Size() / 2
	filePath := logFile.Name()

	logFile.Close()

	readFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer readFile.Close()

	_, err = readFile.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	remaining, err := io.ReadAll(readFile)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, remaining, 0644)
	if err != nil {
		return err
	}

	logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logger = log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags)

	return nil
}
