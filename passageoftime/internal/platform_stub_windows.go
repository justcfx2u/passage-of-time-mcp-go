//go:build windows
// +build windows

package internal

import "errors"

// GetSystemTimezoneUnix is a stub for Windows platform
func GetSystemTimezoneUnix() (string, error) {
	return "", errors.New("Unix timezone detection not supported on Windows")
}
