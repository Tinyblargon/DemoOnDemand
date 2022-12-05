package provider

import (
	"fmt"
	"time"
)

// defaultAPITimeout is a default timeout value that is passed to functions
var defaultAPITimeout time.Duration

// Sets the default api timeout
func Initialize(apiTimeoutInSeconds uint) error {
	if apiTimeoutInSeconds == 0 {
		return fmt.Errorf("defaultAPITimeout not be 0 seconds")
	}
	if defaultAPITimeout != 0 {
		return fmt.Errorf("defaultAPITimeout can only be set once")
	}
	defaultAPITimeout = time.Duration(apiTimeoutInSeconds) * time.Second
	return nil
}

func GetTimeout() time.Duration {
	return defaultAPITimeout
}
