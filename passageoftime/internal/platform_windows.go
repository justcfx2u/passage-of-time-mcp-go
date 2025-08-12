//go:build windows
// +build windows

// BSD 3-Clause License
//
// Copyright (c) 2021, Carlos Henrique Guardão Gandarez
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its
//    contributors may be used to endorse or promote products derived from
//    this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// Copied from https://github.com/JJCinAZ/go-timezone

package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// GetSystemTimezoneWindows tries to find the local timezone configuration on Windows.
// Windows is special. It has unique time zone names (in several meanings of the word) available,
// but unfortunately, they can be translated to the language of the operating system,
// so we need to do a backwards lookup, by going through all time zones and see which
// one matches.
func GetSystemTimezoneWindows() (string, error) {
	// first try the ENV setting
	if tzenv := parseEnvWindows(); tzenv != "" {
		return tzenv, nil
	}

	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\TimeZoneInformation`,
		registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("failed to open registry key: %w", err)
	}

	defer key.Close()

	tzwin, _, err := key.GetStringValue("TimeZoneKeyName")
	if err != nil {
		return "", fmt.Errorf("can not find windows timezone configuration: %w", err)
	}

	// for some reason this returns a string with loads of NULL bytes at
	// least on some systems. I don't know if this is a bug somewhere, I
	// just work around it
	tzwin = strings.ReplaceAll(tzwin, "\x00", "")

	// Use our CLDR-generated mapping instead of the old hardcoded one
	tz, ok := GetIANATimezone(tzwin)
	if !ok {
		// try adding "Standard Time", it seems to work a lot of times
		tzwin += " Standard Time"
		tz, ok = GetIANATimezone(tzwin)
	}

	if !ok {
		return "", fmt.Errorf("windows timezone '%s' not found in CLDR mapping", tzwin)
	}

	return tz, nil
}

// parseEnvWindows parses timezone from TZ env var on Windows.
func parseEnvWindows() string {
	tzenv := os.Getenv("TZ")
	if tzenv == "" {
		return ""
	}

	// On Windows, if TZ is set, it could be:
	// 1. An IANA timezone name (preferred)
	// 2. A Windows timezone name (needs mapping)
	// 3. A file path (unusual on Windows)

	// First, check if it's already a valid IANA timezone by looking for the slash
	if strings.Contains(tzenv, "/") {
		// Looks like IANA format (e.g., "America/New_York")
		return tzenv
	}

	// Try to map it as a Windows timezone name using our CLDR data
	if iana, exists := GetIANATimezone(tzenv); exists {
		return iana
	}

	if filepath.IsAbs(tzenv) && fileExistsWindows(tzenv) {
		// it's a file specification (unusual on Windows but handle it)
		parts := strings.Split(tzenv, string(os.PathSeparator))

		// is it a zone info zone?
		joined := strings.Join(parts[len(parts)-2:], "/")
		if iana, exists := GetIANATimezone(joined); exists {
			return iana
		}

		// maybe it's a short one, like UTC?
		if iana, exists := GetIANATimezone(parts[len(parts)-1]); exists {
			return iana
		}
	}

	// If all else fails, return it as-is and hope for the best
	return tzenv
}

// fileExistsWindows checks if a file or directory exist on Windows.
func fileExistsWindows(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}
