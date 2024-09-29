package mode

import (
	"github.com/gin-gonic/gin"
)

const (
	// Debug debug mode
	Debug = "debug"
	// Release release mode
	Release = "release"
	// Test test mode
	Test = "test"
)

var appMode = Release

// Set sets the new mode.
func Set(newMode string) {
	appMode = newMode
	updateGinMode()
}

// Get returns the current mode.
func Get() string {
	return appMode
}

func updateGinMode() {
	switch Get() {
	case Debug:
		gin.SetMode(gin.DebugMode)
	case Test:
		gin.SetMode(gin.TestMode)
	case Release:
		gin.SetMode(gin.ReleaseMode)
	default:
		panic("unknown mode")
	}
}

// IsDev returns true if the current mode is dev mode.
func IsDev() bool {
	return Get() == Debug || Get() == Test
}
