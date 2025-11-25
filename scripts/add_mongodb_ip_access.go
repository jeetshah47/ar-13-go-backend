// add_mongodb_ip_access.go - Adds current IP address to MongoDB Atlas IP access list
//
// This script:
//   - Gets the current public IP address
//   - Adds it to MongoDB Atlas IP access list using the Atlas Admin API
//
// Usage:
//
//	go run scripts/add_mongodb_ip_access.go
//	go run scripts/add_mongodb_ip_access.go -ip "192.168.1.1"  # Add specific IP
//	go run scripts/add_mongodb_ip_access.go -comment "Development machine"  # Add comment
//
// Flags:
//
//	-ip string        Specific IP address to add (if not provided, current public IP will be detected)
//	-comment string   Comment/description for the IP entry (default: "Added via script")
//	-project-id string MongoDB Atlas Project ID (required)
//	-public-key string MongoDB Atlas Public API Key (required)
//	-private-key string MongoDB Atlas Private API Key (required)
//
// Environment Variables Required:
//
//	MONGODB_ATLAS_PUBLIC_KEY   MongoDB Atlas Public API Key
//	MONGODB_ATLAS_PRIVATE_KEY  MongoDB Atlas Private API Key
//	MONGODB_ATLAS_PROJECT_ID   MongoDB Atlas Project ID
//
// Note: You can create API keys in MongoDB Atlas:
//  1. Go to https://cloud.mongodb.com
//  2. Navigate to Access Manager > API Keys
//  3. Create a new API key with "Project Owner" or "Organization Owner" permissions
//  4. Copy the Public Key and Private Key
//  5. Get your Project ID from Project Settings > General
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	atlasAPIBaseURL = "https://cloud.mongodb.com/api/atlas/v1.0"
	ipifyURL        = "https://api.ipify.org?format=text"
)

type IPAccessEntry struct {
	IPAddress string `json:"ipAddress"`
	Comment   string `json:"comment,omitempty"`
}

type IPAccessListResponse struct {
	Results []struct {
		IPAddress string `json:"ipAddress"`
		Comment   string `json:"comment,omitempty"`
		CIDRBlock string `json:"cidrBlock,omitempty"`
	} `json:"results"`
	TotalCount int `json:"totalCount"`
}

func main() {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	// Parse command-line flags
	ipFlag := flag.String("ip", "", "Specific IP address to add (if not provided, current public IP will be detected)")
	comment := flag.String("comment", "Added via script", "Comment/description for the IP entry")
	projectID := flag.String("project-id", "", "MongoDB Atlas Project ID")
	publicKey := flag.String("public-key", "", "MongoDB Atlas Public API Key")
	privateKey := flag.String("private-key", "", "MongoDB Atlas Private API Key")
	flag.Parse()

	// Get credentials from flags or environment variables
	atlasPublicKey := *publicKey
	if atlasPublicKey == "" {
		atlasPublicKey = os.Getenv("MONGODB_ATLAS_PUBLIC_KEY")
	}
	// Trim whitespace (common issue with .env files)
	atlasPublicKey = strings.TrimSpace(atlasPublicKey)
	if atlasPublicKey == "" {
		log.Fatal("MongoDB Atlas Public API Key is required. Set MONGODB_ATLAS_PUBLIC_KEY environment variable or use -public-key flag")
	}

	atlasPrivateKey := *privateKey
	if atlasPrivateKey == "" {
		atlasPrivateKey = os.Getenv("MONGODB_ATLAS_PRIVATE_KEY")
	}
	// Trim whitespace (common issue with .env files)
	atlasPrivateKey = strings.TrimSpace(atlasPrivateKey)
	if atlasPrivateKey == "" {
		log.Fatal("MongoDB Atlas Private API Key is required. Set MONGODB_ATLAS_PRIVATE_KEY environment variable or use -private-key flag")
	}

	atlasProjectID := *projectID
	if atlasProjectID == "" {
		atlasProjectID = os.Getenv("MONGODB_ATLAS_PROJECT_ID")
	}
	// Trim whitespace (common issue with .env files)
	atlasProjectID = strings.TrimSpace(atlasProjectID)
	if atlasProjectID == "" {
		log.Fatal("MongoDB Atlas Project ID is required. Set MONGODB_ATLAS_PROJECT_ID environment variable or use -project-id flag")
	}

	// Debug: Show first/last few characters of keys (for verification without exposing full keys)
	if len(atlasPublicKey) > 8 {
		log.Printf("Using Public Key: %s...%s (length: %d)",
			atlasPublicKey[:4],
			atlasPublicKey[len(atlasPublicKey)-4:],
			len(atlasPublicKey))
	} else {
		log.Printf("Using Public Key: (length: %d)", len(atlasPublicKey))
	}
	log.Printf("Using Project ID: %s", atlasProjectID)

	// Get IP address
	var ipAddress string
	if *ipFlag != "" {
		ipAddress = *ipFlag
		log.Printf("Using provided IP address: %s", ipAddress)
	} else {
		log.Println("Detecting current public IP address...")
		var err error
		ipAddress, err = getCurrentPublicIP()
		if err != nil {
			log.Fatalf("Failed to get current public IP: %v", err)
		}
		log.Printf("Detected public IP address: %s", ipAddress)
	}

	// Check if IP is already in the access list
	log.Println("Checking if IP is already in access list...")
	exists, err := checkIPExists(atlasPublicKey, atlasPrivateKey, atlasProjectID, ipAddress)
	if err != nil {
		log.Printf("Warning: Could not check existing IPs: %v", err)
	} else if exists {
		log.Printf("IP address %s is already in the access list. Skipping...", ipAddress)
		return
	}

	// Add IP to access list
	log.Printf("Adding IP address %s to MongoDB Atlas access list...", ipAddress)
	if err := addIPToAccessList(atlasPublicKey, atlasPrivateKey, atlasProjectID, ipAddress, *comment); err != nil {
		log.Fatalf("Failed to add IP to access list: %v", err)
	}

	log.Printf("Successfully added IP address %s to MongoDB Atlas access list!", ipAddress)
}

// getCurrentPublicIP gets the current public IP address using ipify.org
func getCurrentPublicIP() (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(ipifyURL)
	if err != nil {
		return "", fmt.Errorf("failed to get IP from ipify: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ipify returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	ip := string(bytes.TrimSpace(body))
	if ip == "" {
		return "", fmt.Errorf("received empty IP address")
	}

	return ip, nil
}

// checkIPExists checks if an IP address already exists in the access list
func checkIPExists(publicKey, privateKey, projectID, ipAddress string) (bool, error) {
	url := fmt.Sprintf("%s/groups/%s/accessList", atlasAPIBaseURL, projectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	// Set authentication header
	auth := base64.StdEncoding.EncodeToString([]byte(publicKey + ":" + privateKey))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var accessList IPAccessListResponse
	if err := json.NewDecoder(resp.Body).Decode(&accessList); err != nil {
		return false, err
	}

	// Check if IP exists (exact match or CIDR match)
	for _, entry := range accessList.Results {
		if entry.IPAddress == ipAddress {
			return true, nil
		}
		// Also check if IP is within a CIDR block (basic check)
		if entry.CIDRBlock != "" && entry.CIDRBlock == ipAddress+"/32" {
			return true, nil
		}
	}

	return false, nil
}

// addIPToAccessList adds an IP address to MongoDB Atlas IP access list
func addIPToAccessList(publicKey, privateKey, projectID, ipAddress, comment string) error {
	url := fmt.Sprintf("%s/groups/%s/accessList", atlasAPIBaseURL, projectID)

	entry := IPAccessEntry{
		IPAddress: ipAddress,
		Comment:   comment,
	}

	jsonData, err := json.Marshal([]IPAccessEntry{entry})
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header (Basic Auth with public:private key)
	auth := base64.StdEncoding.EncodeToString([]byte(publicKey + ":" + privateKey))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		// Provide more helpful error messages
		var errorMsg string
		if resp.StatusCode == 401 {
			errorMsg = "Authentication failed. Please verify:\n" +
				"  1. Your Public API Key is correct\n" +
				"  2. Your Private API Key is correct (no extra spaces)\n" +
				"  3. The API key has 'Project Owner' or 'Organization Owner' permissions\n" +
				"  4. The API key hasn't been deleted or disabled"
		} else if resp.StatusCode == 403 {
			errorMsg = "Access forbidden. Please verify:\n" +
				"  1. Your API key has sufficient permissions (Project Owner or Organization Owner)\n" +
				"  2. You're using the correct Project ID"
		} else {
			errorMsg = fmt.Sprintf("API error: %s", string(body))
		}
		return fmt.Errorf("API returned status %d: %s\n%s", resp.StatusCode, string(body), errorMsg)
	}

	return nil
}
