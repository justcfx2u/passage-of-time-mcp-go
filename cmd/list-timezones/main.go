package main

import (
	"fmt"
	"sort"
	"time"
)

// Import the timezone data from the parent package
// This uses the same GetAllTimezoneIDs() data but as a local function
// to avoid circular import issues with the main package

//go:generate go run ../generate-timezone-list -output timezones.go

func main() {
	allTimezones := getAllTimezoneIDs()

	valid := make([]string, 0)

	for _, tz := range allTimezones {
		_, err := time.LoadLocation(tz)
		if err == nil {
			valid = append(valid, tz)
		}
	}

	sort.Strings(valid)

	fmt.Printf("Available timezones (%d total):\n\n", len(valid))
	for i, tz := range valid {
		fmt.Printf("%2d. %s\n", i+1, tz)
	}
}
