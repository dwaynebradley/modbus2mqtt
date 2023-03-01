package logging

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type LoggingLevel int

const (
	TraceLevel LoggingLevel = iota
	DebugLevel
	InfoLevel
	WarningLevel
	ErrorLevel
	FatalLevel
)

var globalLoggingLevel LoggingLevel = ErrorLevel
var globalLoggingLevelText string = "ERROR"

func SetLoggingLevel(level string) {
	uLevel := strings.ToUpper(level)

	switch uLevel {
	case "TRACE":
		globalLoggingLevel = TraceLevel
	case "DEBUG":
		globalLoggingLevel = DebugLevel
	case "INFO":
		globalLoggingLevel = InfoLevel
	case "WARN":
		globalLoggingLevel = WarningLevel
	case "ERROR":
		globalLoggingLevel = ErrorLevel
	case "FATAL":
		globalLoggingLevel = FatalLevel
	default:
		globalLoggingLevel = ErrorLevel
		uLevel = "ERROR"
	}

	globalLoggingLevelText = uLevel

}

func GetLoggingLevel() LoggingLevel {
	return globalLoggingLevel
}

func NewFieldMap(k string, v string) map[string]string {
	field := make(map[string]string)

	field[k] = v

	return field
}

func AddField(currentFields map[string]string, k string, v string) map[string]string {
	currentFields[k] = v

	return currentFields
}

func Trace(message string) {
	print(TraceLevel, message)
}

func Tracef(message string, fieldMap map[string]string) {
	printWithFields(TraceLevel, message, fieldMap)
}

func Debug(message string) {
	print(DebugLevel, message)
}

func Debugf(message string, fieldMap map[string]string) {
	printWithFields(DebugLevel, message, fieldMap)
}

func Info(message string) {
	print(InfoLevel, message)
}

func Infof(message string, fieldMap map[string]string) {
	printWithFields(InfoLevel, message, fieldMap)
}

func Warning(message string) {
	print(WarningLevel, message)
}

func Warningf(message string, fieldMap map[string]string) {
	printWithFields(WarningLevel, message, fieldMap)
}

func Error(message string) {
	print(ErrorLevel, message)
}

func Errorf(message string, fieldMap map[string]string) {
	printWithFields(ErrorLevel, message, fieldMap)
}

func Fatal(message string) {
	print(FatalLevel, message)
	os.Exit(125)
}

func Fatalf(message string, fieldMap map[string]string) {
	printWithFields(FatalLevel, message, fieldMap)
	os.Exit(125)
}

func print(level LoggingLevel, message string) {
	printWithFields(level, message, make(map[string]string))
}

func printWithFields(level LoggingLevel, message string, fieldMap map[string]string) {
	if level >= globalLoggingLevel {
		log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Lmicroseconds)

		formattedMessage := "level=" + globalLoggingLevelText + " msg=" + strconv.Quote(message)

		if len(fieldMap) > 0 {
			for k, v := range fieldMap {
				formattedMessage += " " + k + "=" + strconv.Quote(v)
			}
		}

		log.Print(formattedMessage)
	}
}
