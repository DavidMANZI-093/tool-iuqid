package logger

import (
	"log"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
)

func Info(format string, v ...interface{}) {
	log.Printf(ColorCyan+"(info) "+ColorReset+format, v...)
}

func Success(format string, v ...interface{}) {
	log.Printf(ColorGreen+"(success) "+ColorReset+format, v...)
}

func Warn(format string, v ...interface{}) {
	log.Printf(ColorYellow+"(warn) "+ColorReset+format, v...)
}

func Error(format string, v ...interface{}) {
	log.Printf(ColorRed+"(error) "+ColorReset+format, v...)
}
