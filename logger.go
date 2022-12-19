package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
)

func init() {
	debugLog = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix)
	infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix)
	warningLog = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix)
	errorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix)

}

const (
	DebugLevel   = 1
	InfoLevel    = 0
	WarningLevel = -1
	ErrorLevel   = -2
	Disabled     = -3
)

var (
	infoLogChannel    = make(chan string, 2<<6)
	debugLogChannel   = make(chan string, 2<<8)
	warningLogChannel = make(chan string, 2<<6)
	errorLogChannel   = make(chan string, 2<<6)

	debugLog   *log.Logger
	infoLog    *log.Logger
	warningLog *log.Logger
	errorLog   *log.Logger

	logLevel = InfoLevel
)

// listenLogChannel adds all elements of the string slice into a byte buffer,
// afterwards the content of the buffer will be written
// into the specified logger.
func listenLogChannel(logWriter *log.Logger, channel <-chan string) {
	for msg := range channel {
		logWriter.Println(msg)
	}
}

// InitNewLogger sets the io.Writer interface as an output for
// the loggers and spawns a goroutine for each of the
// available loggers.
func InitNewLogger(outInterface io.Writer, level int) {
	logLevel = level

	debugLog.SetOutput(outInterface)
	infoLog.SetOutput(outInterface)
	warningLog.SetOutput(outInterface)
	errorLog.SetOutput(outInterface)

	go listenLogChannel(debugLog, debugLogChannel)
	go listenLogChannel(errorLog, errorLogChannel)
	go listenLogChannel(warningLog, warningLogChannel)
	go listenLogChannel(infoLog, infoLogChannel)
}

// Debug writes a message to the debug log in a
// synchronous manner, if debug level enabled.
func Debug(message ...interface{}) {
	if DebugLevel <= logLevel {
		debugLog.Println(parseMsg(message))
	}
}

// DebugAsync sends the given strings to the debug log channel.
func DebugAsync(message ...interface{}) {
	if DebugLevel <= logLevel {
		debugLogChannel <- parseMsg(message)
	}
}

// Info writes a message to the info log in a
// synchronous manner, if info level enabled.
func Info(message ...interface{}) {
	if InfoLevel <= logLevel {
		infoLog.Println(parseMsg(message))
	}
}

// InfoAsync sends the given string arguments to
// the info log channel.
func InfoAsync(message ...interface{}) {
	if InfoLevel <= logLevel {
		infoLogChannel <- parseMsg(message)
	}
}

// Warning writes a message to the warning log
// in a synchronous manner, if warning level enabled.
func Warning(message ...interface{}) {
	if WarningLevel <= logLevel {
		warningLog.Println(parseMsg(message))
	}
}

// WarningAsync sends the given string arguments to the
// warning log channel.
func WarningAsync(message ...interface{}) {
	if WarningLevel <= logLevel {
		warningLogChannel <- parseMsg(message)
	}
}

// Error writes a message to the error log
// in a synchronous manner, if error level enabled.
func Error(message ...interface{}) {
	if ErrorLevel <= logLevel {
		errorLog.Println(parseMsg(message))
	}
}

// ErrorAsync sends the given string arguments to the
// error log channel.
func ErrorAsync(message ...interface{}) {
	if ErrorLevel <= logLevel {
		errorLogChannel <- parseMsg(message)
	}
}

// parseMsg turns all given variables into strings and
// return the message as a string.
func parseMsg(vars ...interface{}) string {
	var msg string

	for _, v := range vars {
		txt := fmt.Sprintf("%v", v)
		txt = strings.Trim(txt, "[")
		txt = strings.Trim(txt, "]")
		msg += fmt.Sprintf("%v", txt)
	}

	function, file, line, _ := runtime.Caller(2)
	msg = fmt.Sprintf("%v [%v => %v:%v]", msg, path.Base(file), path.Base(runtime.FuncForPC(function).Name()), line)

	return msg
}

// GetLogLevelID determines what log verbosity to use.
// level - usually this is a client input.
func GetLogLevelID(level string) int {
	var selectedLevel = InfoLevel

	if strings.EqualFold("debug", level) {
		selectedLevel = DebugLevel
	} else if strings.EqualFold("warning", level) {
		selectedLevel = WarningLevel
	} else if strings.EqualFold("error", level) {
		selectedLevel = ErrorLevel
	} else if strings.EqualFold("quiet", level) {
		selectedLevel = Disabled
	}

	return selectedLevel
}

// GetLogLevel returns the current log level
func GetLogLevel() int {
	return logLevel
}

// SetLogLevel sets the log level to output.
func SetLogLevel(level int) {
	logLevel = level
}
