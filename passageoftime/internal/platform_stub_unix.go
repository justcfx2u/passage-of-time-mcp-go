//go:build darwin || linux
// +build darwin linux

package internal

import "errors"

// GetSystemTimezoneWindows is a stub for Unix platforms
func GetSystemTimezoneWindows() (string, error) {
	return "", errors.New("Windows timezone detection not supported on Unix platforms")
}
