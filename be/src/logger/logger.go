package logger

import (
	"fmt"
	"log"
	"os"
)

func Debug(format string, args ...any) {
	log.Printf("DEBUG %s", fmt.Sprintf(format, args...))
}

func Trace(format string, args ...any) {
	log.Printf("TRACE %s", fmt.Sprintf(format, args...))
}

func Info(format string, args ...any) {
	log.Printf("INFO %s", fmt.Sprintf(format, args...))
}

func Error(format string, args ...any) {
	log.Printf("ERROR %s", fmt.Sprintf(format, args...))
}

func Fatal(format string, args ...any) {
	log.Printf("FATAL %s", fmt.Sprintf(format, args...))
	os.Exit(1)
}

