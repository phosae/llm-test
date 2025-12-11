package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	// Default Cloud Platform scope
	cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
)

type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

type ADCConfig struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

func main() {
	// Read ADC file path from command line argument or use default
	adcFilePath := "adc.json"
	if len(os.Args) > 1 {
		adcFilePath = os.Args[1]
	}

	if _, err := os.Stat(adcFilePath); os.IsNotExist(err) {
		panic(fmt.Sprintf("ADC file %s does not exist, gen-vetexai-token <adc-file-path>", adcFilePath))
	}

	// Read ADC JSON from file
	adcData, err := os.ReadFile(adcFilePath)
	if err != nil {
		log.Fatalf("Failed to read ADC file %s: %v", adcFilePath, err)
	}

	// Parse ADC config for display
	var adcConfig ADCConfig
	if err := json.Unmarshal(adcData, &adcConfig); err != nil {
		log.Printf("Warning: Could not parse ADC config for display: %v", err)
	}

	// Get token (either from cache or generate new one)
	token, err := getOrGenerateToken(adcData, adcConfig.ProjectID)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	// Beautify and display token information
	fmt.Println("🔐 Google Cloud Platform Token Information")
	fmt.Println("==========================================")
	fmt.Println()

	// Display ADC configuration
	fmt.Println("📋 ADC Configuration:")
	fmt.Printf("   Project ID:     %s\n", adcConfig.ProjectID)
	fmt.Printf("   Client Email:   %s\n", adcConfig.ClientEmail)
	fmt.Printf("   Private Key ID: %s\n", adcConfig.PrivateKeyID)
	fmt.Printf("   Type:           %s\n", adcConfig.Type)
	fmt.Println()

	// Display token information
	fmt.Println("🎫 Token Details:")
	fmt.Printf("   Access Token:   %s...%s\n",
		token.AccessToken[:20],
		token.AccessToken[len(token.AccessToken)-10:])
	fmt.Printf("   Token Type:     %s\n", token.TokenType)

	if !token.Expiry.IsZero() {
		fmt.Printf("   Expires At:     %s\n", token.Expiry.Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("   Time Remaining: %s\n", time.Until(token.Expiry))
	} else {
		fmt.Printf("   Expires At:     Never\n")
	}

	if token.RefreshToken != "" {
		fmt.Printf("   Refresh Token:  %s...%s\n",
			token.RefreshToken[:20],
			token.RefreshToken[len(token.RefreshToken)-10:])
	}
	fmt.Println()

	// Display scope information
	fmt.Println("🔑 Scope Information:")
	fmt.Printf("   Scope:          %s\n", cloudPlatformScope)
	fmt.Println()

	// Display full token
	fmt.Println("🔍 Full Access Token:")
	fmt.Printf("%s", token.AccessToken)
	fmt.Println()
}

// getOrGenerateToken retrieves a cached token if valid, or generates a new one
func getOrGenerateToken(adcData []byte, projectID string) (*oauth2.Token, error) {
	// Get token cache directory
	cacheDir := filepath.Join(os.Getenv("HOME"), ".llm-jwt-tokens")
	tokenFile := filepath.Join(cacheDir, fmt.Sprintf("vertexai_%s", projectID))

	// Try to read cached token
	if cachedToken, err := readCachedToken(tokenFile); err == nil && cachedToken != nil {
		// Check if token is still valid (with 5 minute buffer)
		if !cachedToken.Expiry.IsZero() && time.Until(cachedToken.Expiry) > 5*time.Minute {
			fmt.Println("✅ Using cached token (not expired)")
			return cachedToken, nil
		}
		fmt.Println("⚠️  Cached token expired, generating new one...")
	} else {
		fmt.Println("📝 No cached token found, generating new one...")
	}

	// Generate new token
	token, err := generateNewToken(adcData)
	if err != nil {
		return nil, err
	}

	// Cache the new token
	if err := cacheToken(tokenFile, token); err != nil {
		fmt.Printf("⚠️  Warning: Failed to cache token: %v\n", err)
	}

	return token, nil
}

// readCachedToken reads a token from the cache file
func readCachedToken(tokenFile string) (*oauth2.Token, error) {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// cacheToken saves a token to the cache file
func cacheToken(tokenFile string, token *oauth2.Token) error {
	// Create cache directory if it doesn't exist
	cacheDir := filepath.Dir(tokenFile)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Marshal token to JSON
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(tokenFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	fmt.Printf("💾 Token cached to: %s\n", tokenFile)
	return nil
}

// generateNewToken creates a new token from ADC data
func generateNewToken(adcData []byte) (*oauth2.Token, error) {
	// Create JWT config from ADC JSON
	jwtConfig, err := google.JWTConfigFromJSON(adcData, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT config: %w", err)
	}

	// Get token source
	ts := jwtConfig.TokenSource(context.Background())

	// Get token
	token, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return token, nil
}
