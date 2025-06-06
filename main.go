package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func main() {
	// Define command-line flags
	var (
		testTypes = flag.String("test", "", "Test types: f(unction), v(ision). Use comma-separated for multiple: f,v")
		showHelp  = flag.Bool("h", false, "Show help")
	)

	flag.Parse()

	// Handle help
	if *showHelp {
		fmt.Println("Usage: go run . [flags]")
		fmt.Println("\nFlags:")
		fmt.Println("  -test string    Test types: f(unction), v(ision)")
		fmt.Println("                  Examples: -test f    (function only)")
		fmt.Println("                           -test v    (vision only)")
		fmt.Println("                           -test f,v  (both)")
		fmt.Println("  -h              Show this help message")
		fmt.Println("\nDefault behavior (no -test flag): Test both function and vision")
		fmt.Println("\nEnvironment Variables:")
		fmt.Println("  API_KEY     - Your API key")
		fmt.Println("  BASE_URL    - Custom base URL (optional)")
		fmt.Println("  MODEL       - Model to use (e.g., gpt-4-vision-preview)")
		return
	}

	// Parse test types
	var shouldTestFunction, shouldTestVision bool

	if *testTypes == "" {
		// Default: test both
		shouldTestFunction = true
		shouldTestVision = true
	} else {
		// Parse comma-separated values
		types := strings.Split(strings.ToLower(*testTypes), ",")
		for _, t := range types {
			t = strings.TrimSpace(t)
			switch t {
			case "f", "function":
				shouldTestFunction = true
			case "v", "vision":
				shouldTestVision = true
			default:
				fmt.Printf("Unknown test type: %s\n", t)
				fmt.Println("Valid types: f(unction), v(ision)")
				return
			}
		}
	}

	ctx := context.Background()
	cfg := openai.DefaultConfig(os.Getenv("API_KEY"))
	cfg.BaseURL = os.Getenv("BASE_URL")
	client := openai.NewClientWithConfig(cfg)

	// Test function calling
	if shouldTestFunction {
		fmt.Println("🔧 Testing Function Calling...")
		function(ctx, client)
	}

	// Test vision API
	if shouldTestVision {
		if shouldTestFunction {
			fmt.Println("\n" + strings.Repeat("=", 50))
		}
		fmt.Println("👁️  Testing Vision API...")
		vision(ctx, client)
	}
}
