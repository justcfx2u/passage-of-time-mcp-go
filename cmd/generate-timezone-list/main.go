package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

func main() {
	var outputFile = flag.String("output", "", "Output file for generated timezone function")
	flag.Parse()
	
	fmt.Println("🌍 Go Timezone Database Extractor")
	fmt.Println("===================================")
	
	// Extract timezone data from Go's zoneinfo.zip (shared logic)
	timezones, err := extractTimezonesFromGo()
	if err != nil {
		fmt.Printf("❌ Failed to extract timezones: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("✅ Extracted %d timezone identifiers from Go's timezone database\n", len(timezones))
	
	// Validate a sample of timezones to ensure they work with Go's time package (shared logic)
	err = validateTimezones(timezones)
	if err != nil {
		fmt.Printf("❌ Timezone validation failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✅ Timezone validation passed")
	
	// Branch based on output mode
	if *outputFile != "" {
		// Generate standalone function for go:generate
		err := generateStandaloneFunction(*outputFile, timezones)
		if err != nil {
			fmt.Printf("❌ Failed to generate standalone function: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Generated standalone timezone function in %s\n", *outputFile)
		return
	}
	
	// Generate the popular timezones subset (time_utils.go mode only)
	popularTimezones := generatePopularTimezones(timezones)
	fmt.Printf("✅ Generated %d popular timezones from extracted data\n", len(popularTimezones))
	
	// Generate the Go source file (time_utils.go mode)
	fmt.Println("\n📝 Generating time_utils.go...")
	err = generateTimeUtilsFile(timezones, popularTimezones)
	if err != nil {
		fmt.Printf("❌ Failed to generate Go file: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✅ Successfully generated automated timezone lists in time_utils.go")
}

// extractTimezonesFromGo extracts timezone identifiers from Go's embedded zoneinfo.zip
func extractTimezonesFromGo() ([]string, error) {
	// Get Go's zoneinfo.zip path
	goroot := runtime.GOROOT()
	zoneinfoPath := filepath.Join(goroot, "lib", "time", "zoneinfo.zip")
	
	fmt.Printf("📁 Reading timezone data from: %s\n", zoneinfoPath)
	
	// Open the zip file
	reader, err := zip.OpenReader(zoneinfoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zoneinfo.zip: %w", err)
	}
	defer reader.Close()
	
	var timezones []string
	timezoneSet := make(map[string]bool)
	
	// Process each file in the zip
	for _, file := range reader.File {
		// Skip directories and non-timezone files
		if file.FileInfo().IsDir() || strings.Contains(file.Name, "..") {
			continue
		}
		
		// Skip special files that aren't timezone identifiers
		if strings.HasPrefix(file.Name, "iso3166.tab") ||
		   strings.HasPrefix(file.Name, "leapseconds") ||
		   strings.HasPrefix(file.Name, "zone.tab") ||
		   strings.HasPrefix(file.Name, "zone1970.tab") {
			continue
		}
		
		// Clean the timezone identifier
		timezone := strings.TrimSpace(file.Name)
		if timezone != "" && !timezoneSet[timezone] {
			timezoneSet[timezone] = true
			timezones = append(timezones, timezone)
		}
	}
	
	// Sort for consistent output
	sort.Strings(timezones)
	
	return timezones, nil
}

// validateTimezones validates that a sample of extracted timezones work with Go's time package
func validateTimezones(timezones []string) error {
	fmt.Println("🔍 Validating extracted timezones...")
	
	// Test a representative sample
	sampleSize := 20
	if len(timezones) < sampleSize {
		sampleSize = len(timezones)
	}
	
	step := len(timezones) / sampleSize
	if step < 1 {
		step = 1
	}
	
	validCount := 0
	for i := 0; i < len(timezones); i += step {
		timezone := timezones[i]
		_, err := time.LoadLocation(timezone)
		if err == nil {
			validCount++
		} else {
			fmt.Printf("⚠️  Invalid timezone detected: %s (%v)\n", timezone, err)
		}
		
		if validCount >= sampleSize {
			break
		}
	}
	
	fmt.Printf("✅ Validated %d/%d sample timezones\n", validCount, sampleSize)
	
	if validCount == 0 {
		return fmt.Errorf("no valid timezones found in sample")
	}
	
	return nil
}

// generatePopularTimezones creates a curated list of popular timezones from the full list
func generatePopularTimezones(allTimezones []string) []string {
	// Define patterns for popular timezone regions and cities
	popularPatterns := []string{
		// UTC
		"UTC",
		
		// Americas - Major business hubs
		"America/New_York",      // Eastern US
		"America/Chicago",       // Central US  
		"America/Denver",        // Mountain US
		"America/Los_Angeles",   // Pacific US
		"America/Toronto",       // Eastern Canada
		"America/Mexico_City",   // Mexico
		"America/Sao_Paulo",     // Brazil
		"America/Buenos_Aires",  // Argentina
		
		// Europe - Major business hubs
		"Europe/London",         // UK
		"Europe/Paris",          // France
		"Europe/Berlin",         // Germany
		"Europe/Rome",           // Italy
		"Europe/Amsterdam",      // Netherlands
		"Europe/Zurich",         // Switzerland
		"Europe/Stockholm",      // Sweden
		"Europe/Moscow",         // Russia
		
		// Asia - Major business hubs
		"Asia/Tokyo",            // Japan
		"Asia/Shanghai",         // China
		"Asia/Hong_Kong",        // Hong Kong
		"Asia/Singapore",        // Singapore
		"Asia/Kolkata",          // India
		"Asia/Dubai",            // UAE
		
		// Australia & Pacific
		"Australia/Sydney",      // Eastern Australia
		"Pacific/Auckland",      // New Zealand
	}
	
	// Build the popular list from available timezones
	var popularTimezones []string
	timezoneSet := make(map[string]bool)
	
	// Convert all timezones to a set for quick lookup
	for _, tz := range allTimezones {
		timezoneSet[tz] = true
	}
	
	// Add popular timezones that exist in the extracted list
	for _, pattern := range popularPatterns {
		if timezoneSet[pattern] {
			popularTimezones = append(popularTimezones, pattern)
		}
	}
	
	return popularTimezones
}

// generateTimeUtilsFile generates the time_utils.go file with automated timezone lists
func generateTimeUtilsFile(allTimezones, popularTimezones []string) error {
	// Read the current time_utils.go to preserve non-timezone functions
	currentContent, err := os.ReadFile("time_utils.go")
	if err != nil {
		return fmt.Errorf("failed to read current time_utils.go: %w", err)
	}
	
	// Split content to find where to replace timezone functions
	contentStr := string(currentContent)
	
	// Find the start of getPopularTimezones function
	popularStart := strings.Index(contentStr, "// getPopularTimezones returns the")
	if popularStart == -1 {
		return fmt.Errorf("could not find getPopularTimezones function in time_utils.go")
	}
	
	// Find the start of getAllTimezoneIDs function  
	allStart := strings.Index(contentStr, "// getAllTimezoneIDs returns all available")
	if allStart == -1 {
		return fmt.Errorf("could not find getAllTimezoneIDs function in time_utils.go")
	}
	
	// Find the end of getAllTimezoneIDs function (next function or end of file)
	allEnd := strings.Index(contentStr[allStart:], "\n// formatOffset")
	if allEnd == -1 {
		allEnd = len(contentStr)
	} else {
		allEnd = allStart + allEnd
	}
	
	// Build the new timezone functions
	var newFunctions strings.Builder
	
	// Generate getPopularTimezones function
	newFunctions.WriteString("// getPopularTimezones returns the most commonly used timezones\n")
	newFunctions.WriteString("// Generated automatically from Go's timezone database\n")
	newFunctions.WriteString("// Represents major business/tech hubs across Americas, Europe, Asia, and Australia\n")
	newFunctions.WriteString("func getPopularTimezones() []string {\n")
	newFunctions.WriteString("\treturn []string{\n")
	
	for _, tz := range popularTimezones {
		newFunctions.WriteString(fmt.Sprintf("\t\t%q,\n", tz))
	}
	
	newFunctions.WriteString("\t}\n")
	newFunctions.WriteString("}\n\n")
	
	// Generate getAllTimezoneIDs function
	newFunctions.WriteString("// getAllTimezoneIDs returns all available IANA timezone identifiers\n")
	newFunctions.WriteString("// Generated automatically from Go's embedded timezone database\n")
	newFunctions.WriteString("// Source: $GOROOT/lib/time/zoneinfo.zip\n")
	newFunctions.WriteString(fmt.Sprintf("// Total timezones: %d\n", len(allTimezones)))
	newFunctions.WriteString("func getAllTimezoneIDs() []string {\n")
	newFunctions.WriteString("\treturn []string{\n")
	
	for _, tz := range allTimezones {
		newFunctions.WriteString(fmt.Sprintf("\t\t%q,\n", tz))
	}
	
	newFunctions.WriteString("\t}\n")
	newFunctions.WriteString("}\n\n")
	
	// Build the final content
	var finalContent strings.Builder
	
	// Add content before getPopularTimezones
	finalContent.WriteString(contentStr[:popularStart])
	
	// Add generated functions
	finalContent.WriteString(newFunctions.String())
	
	// Add content after getAllTimezoneIDs
	finalContent.WriteString(contentStr[allEnd:])
	
	// Write the updated file
	err = os.WriteFile("time_utils.go", []byte(finalContent.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated time_utils.go: %w", err)
	}
	
	fmt.Printf("📄 Updated time_utils.go with:\n")
	fmt.Printf("  - %d popular timezones\n", len(popularTimezones))
	fmt.Printf("  - %d total timezone identifiers\n", len(allTimezones))
	
	return nil
}

// generateStandaloneFunction creates a standalone timezones.go file with getAllTimezoneIDs function
func generateStandaloneFunction(outputFile string, timezones []string) error {
	// Use pre-extracted and validated timezone data (no duplication)

	// Generate the standalone Go file content
	var content strings.Builder
	
	content.WriteString("// Code generated by go:generate; DO NOT EDIT.\n")
	content.WriteString("// This file contains timezone data extracted from Go's embedded timezone database\n")
	content.WriteString("// Source: $GOROOT/lib/time/zoneinfo.zip\n")
	content.WriteString(fmt.Sprintf("// Generated on: %s\n", time.Now().Format(time.RFC3339)))
	content.WriteString(fmt.Sprintf("// Total timezones: %d\n\n", len(timezones)))
	
	// Determine package name from output file path
	packageName := "main"
	// Normalize path separators for cross-platform detection
	normalizedPath := filepath.ToSlash(outputFile)
	if strings.Contains(normalizedPath, "internal/") {
		packageName = "internal"
	} else if strings.Contains(normalizedPath, "passageoftime/") && !strings.Contains(normalizedPath, "internal/") {
		packageName = "passageoftime"
	}
	
	content.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	
	// Generate getAllTimezoneIDs function with appropriate visibility
	functionName := "getAllTimezoneIDs"
	if packageName == "internal" {
		functionName = "GetAllTimezoneIDs"
	}
	
	content.WriteString(fmt.Sprintf("// %s returns all available IANA timezone identifiers\n", functionName))
	content.WriteString("// Generated automatically from Go's embedded timezone database\n")
	content.WriteString(fmt.Sprintf("func %s() []string {\n", functionName))
	content.WriteString("\treturn []string{\n")
	
	for _, tz := range timezones {
		content.WriteString(fmt.Sprintf("\t\t%q,\n", tz))
	}
	
	content.WriteString("\t}\n")
	content.WriteString("}\n")

	// Write the generated file
	err := os.WriteFile(outputFile, []byte(content.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", outputFile, err)
	}

	return nil
}