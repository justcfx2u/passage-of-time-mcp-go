package internal

import (
	"runtime"
	"time"
)

// GetSystemTimezone returns the system's local timezone name using platform-specific detection.
// Falls back to "UTC" if the system timezone cannot be determined.
//
// Detection strategy:
// 1. Check if Go's standard detection returns a proper IANA timezone
// 2. Use platform-specific detection methods (registry on Windows, file parsing on Unix)
// 3. Fall back to UTC if all methods fail
func GetSystemTimezone() string {
	// Get the system's local timezone using Go's standard method
	loc := time.Now().Location()
	tzName := loc.String()
	
	// If we get a proper IANA timezone name, use it directly
	// This works well on most Unix systems
	if tzName != "" && tzName != "Local" {
		return tzName
	}
	
	// If we get "Local" or empty, we need platform-specific resolution
	// This is common on Windows where Go returns "Local" instead of IANA name
	var detectedTz string
	var err error
	
	if runtime.GOOS == "windows" {
		detectedTz, err = GetSystemTimezoneWindows()
	} else {
		// Unix-like systems (Linux, macOS, etc.)
		detectedTz, err = GetSystemTimezoneUnix()
	}
	
	if err == nil && detectedTz != "" {
		return detectedTz
	}
	
	// Final fallback to UTC if we can't determine the actual timezone
	return "UTC"
}
