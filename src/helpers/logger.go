package helpers

/*
logger.go contains functions to help logging or debugging.

TODO: add more functionality besides a simple "debugmode"
*/

import (
	"log"
)

const debugMode = true

func DebugLog(format string, v ...any) {
	if debugMode {
		log.Printf("[DEBUG] "+format, v...)
	}
}
