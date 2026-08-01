package helpers

import (
	"log"
)

const debugMode = true

func DebugLog(format string, v ...any) {
	if debugMode {
		log.Printf("[DEBUG] "+format, v...)
	}
}
