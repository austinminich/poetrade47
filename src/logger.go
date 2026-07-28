package main

import (
	"log"
)

const debugMode = true

func debugLog(format string, v ...any) {
	if debugMode {
		log.Printf("[DEBUG] "+format, v...)
	}
}
