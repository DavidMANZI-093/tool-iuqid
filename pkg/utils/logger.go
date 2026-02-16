package utils

import (
	"log"
)

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
)

func LogInfo(format string, v ...interface{}) {
	log.Printf(colorCyan+"(info) "+colorReset+format, v...)
}

func LogSuccess(format string, v ...interface{}) {
	log.Printf(colorGreen+"(success) "+colorReset+format, v...)
}

func LogWarn(format string, v ...interface{}) {
	log.Printf(colorYellow+"(warn) "+colorReset+format, v...)
}

func LogError(format string, v ...interface{}) {
	log.Printf(colorRed+"(error) "+colorReset+format, v...)
}

func LogDebug(format string, v ...interface{}) {
	log.Printf(colorMagenta+"(error) "+colorReset+format, v...)
}
