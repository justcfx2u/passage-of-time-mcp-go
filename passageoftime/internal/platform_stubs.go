//go:build !windows && !darwin && !linux
// +build !windows,!darwin,!linux

package internal

import "errors"

// GetSystemTimezoneWindows is a stub for non-Windows platforms
func GetSystemTimezoneWindows() (string, error) {
	return "", errors.New("Windows timezone detection not supported on this platform")
}

// GetSystemTimezoneUnix is a stub for non-Unix platforms
func GetSystemTimezoneUnix() (string, error) {
	return "", errors.New("Unix timezone detection not supported on this platform")
}
