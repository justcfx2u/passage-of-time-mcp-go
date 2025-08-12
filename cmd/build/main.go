package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("🔨 Building Passage of Time MCP Server - Cross-Platform GitHub Actions Naming")
	
	// Step 1: Generate timezone data and validate
	fmt.Println("\n🌍 Generating timezone data...")
	if err := generateTimezoneData(); err != nil {
		fmt.Printf("❌ Failed to generate timezone data: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Timezone data generation completed")
	
	// Define cross-platform build matrix (matching GitHub Actions)
	platforms := []struct {
		goos   string
		goarch string
	}{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"windows", "amd64"},
		// {"windows", "arm64"}, // Excluded per GitHub Actions
		{"darwin", "amd64"},
		{"darwin", "arm64"},
	}
	
	// Single build variant - semantic differentiation removed as unused
	variant := struct {
		name        string
		description string
	}{
		name:        "passage-of-time-mcp",
		description: "MCP time/date server with 4-layer parsing",
	}
	
	// Generate all build combinations
	builds := []struct {
		output      string
		description string
		goos        string
		goarch      string
	}{}
	
	for _, platform := range platforms {
		ext := ""
		if platform.goos == "windows" {
			ext = ".exe"
		}
		
		builds = append(builds, struct {
			output      string
			description string
			goos        string
			goarch      string
		}{
			output:      fmt.Sprintf("%s-%s-%s%s", variant.name, platform.goos, platform.goarch, ext),
			description: fmt.Sprintf("%s (%s/%s)", variant.description, platform.goos, platform.goarch),
			goos:        platform.goos,
			goarch:      platform.goarch,
		})
	}
	
	// Build each target
	success := 0
	for _, build := range builds {
		fmt.Printf("\n📦 Building binary...\n")
		fmt.Printf("   Description: %s\n", build.description)
		fmt.Printf("   Output: %s\n", build.output)
		
		// Prepare build command
		args := []string{"build"}
		args = append(args, "-ldflags", "-s -w") // Strip debug info for smaller binaries
		args = append(args, "-o", build.output)
		
		// Execute build with cross-platform environment
		cmd := exec.Command("go", args...)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GOOS=%s", build.goos),
			fmt.Sprintf("GOARCH=%s", build.goarch),
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		err := cmd.Run()
		if err != nil {
			fmt.Printf("❌ Failed to build binary: %v\n", err)
			continue
		}
		
		// Check if binary was created
		if _, err := os.Stat(build.output); err != nil {
			fmt.Printf("❌ Binary %s not found after build\n", build.output)
			continue
		}
		
		fmt.Printf("✅ Successfully built %s\n", build.output)
		success++
	}
	
	fmt.Printf("\n🎉 Build Summary: %d/%d builds successful\n", success, len(builds))
	
	if success == len(builds) {
		fmt.Println("\n📋 Usage:")
		fmt.Printf("   Run any binary: ./%s\n", builds[0].output)
		fmt.Println("\n📝 Features include 4-layer parsing:")
		fmt.Println("   - Duration parsing (-14d, 2h30m, -5s)")
		fmt.Println("   - Standard formats (ISO 8601, RFC3339)")
		fmt.Println("   - Natural language (tomorrow, next Monday)")
		fmt.Println("   - Timezone detection and parameter support")
	}
	
	if success != len(builds) {
		os.Exit(1)
	}
}

// generateTimezoneData runs timezone generation and validates extraction
func generateTimezoneData() error {
	// Get the current working directory
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	
	// The working directory should already be the project root when called as `go run cmd/build/main.go`
	// So we use pwd as the project root
	projectRoot := pwd
	
	// Calculate paths relative to project root  
	generatorPath := filepath.Join(projectRoot, "cmd", "generate-timezone-list", "main.go")
	outputPath := filepath.Join(projectRoot, "passageoftime", "internal", "timezones.go")
	
	fmt.Printf("   Project root: %s\n", projectRoot)
	fmt.Printf("   Generator path: %s\n", generatorPath)
	fmt.Printf("   Output path: %s\n", outputPath)
	fmt.Println("   Running timezone generation tool...")
	
	// Run the timezone generation tool for the library's internal timezone data
	cmd := exec.Command("go", "run", generatorPath, "-output", outputPath)
	cmd.Dir = projectRoot // Set working directory to project root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("timezone generation failed: %w", err)
	}
	
	// Validate cmd/list-timezones generation
	fmt.Println("   Generating list-timezones data...")
	cmd = exec.Command("go", "generate", "./cmd/list-timezones")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("list-timezones generation failed: %w", err)
	}
	
	// Validate extraction was successful by checking timezone count
	fmt.Println("   Validating timezone extraction...")
	testPath := filepath.Join(projectRoot, "cmd", "test-timezone-automation", "main.go")
	cmd = exec.Command("go", "run", testPath)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("timezone validation failed: %w", err)
	}
	
	return nil
}