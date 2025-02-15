package logger

import (
	"auth_service/configs"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	ERROR   = "ERROR"
	WARNING = "WARNING"
	EVENT   = "EVENT"
	INFO    = "INFO"
	DEBUG   = "DEBUG"
)

var logLevel string = INFO

// SetLogLevel sets the log level for the logger
func SetLogLevel(level string) {
	logLevel = level
}

// getLogPrefix generates the prefix for log messages
func getLogPrefix() string {
	timestr := time.Now().Format(time.RFC3339Nano)
	// Format timestamp to include nanoseconds
	timestrs := strings.Split(timestr, "+")
	tstr := (timestrs[0] + "000000000")
	timestrs[0] = tstr[:29]
	timestr = strings.Join(timestrs, "+")

	appName := configs.GetVersion()
	version := configs.GetAppName()

	// Retrieve caller information
	funcName := ""
	pc, _, line, ok := runtime.Caller(3)
	if ok {
		function := runtime.FuncForPC(pc)
		funcName = function.Name()
	}

	funcString := ""
	if logLevel == DEBUG {
		funcString = funcName + ":" + strconv.Itoa(line) + " - "
	}

	return timestr + " - " + appName + " - VERSION:" + version + " - " + funcString
}

// Log functions for various levels
func Debug(id string, v ...any) {
	if logLevel == DEBUG {
		prefix := getLogPrefix() + id + " - DEBUG - "
		msgs := fmt.Sprint(v...)
		fmt.Println(prefix + msgs)
	}
}

func Info(id string, v ...any) {
	if logLevel == INFO || logLevel == DEBUG {
		prefix := getLogPrefix() + id + " - INFO - "
		msgs := fmt.Sprint(v...)
		fmt.Println(prefix + msgs)
	}
}

func Warning(id string, v ...any) {
	if logLevel == WARNING || logLevel == INFO || logLevel == DEBUG {
		prefix := getLogPrefix() + id + " - WARNING - "
		msgs := fmt.Sprint(v...)
		fmt.Println(prefix + msgs)
	}
}

func Error(id string, v ...any) {
	if logLevel == ERROR || logLevel == WARNING || logLevel == INFO || logLevel == DEBUG {
		prefix := getLogPrefix() + id + " - ERROR - "
		msgs := fmt.Sprint(v...)
		fmt.Println(prefix + msgs)
	}
}
