package logger

import (
	"fmt"
	"log"
	"os"
)

func Debugf(format string, args ...any) {
	log.Printf("DEBUG %s", fmt.Sprintf(format, args...))
}

func Tracef(format string, args ...any) {
	log.Printf("TRACE %s", fmt.Sprintf(format, args...))
}

func Infof(format string, args ...any) {
	log.Printf("INFO %s", fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...any) {
	log.Printf("ERROR %s", fmt.Sprintf(format, args...))
}

func Fatalf(format string, args ...any) {
	log.Printf("FATAL %s", fmt.Sprintf(format, args...))
	os.Exit(1)
}

