package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	endpoint        = flag.String("endpoint", "", "MinIO endpoint (e.g., nas-ip:32774)")
	accessKeyID     = flag.String("access-key", "", "MinIO root access key (MINIO_ROOT_USER)")
	secretAccessKey = flag.String("secret-key", "", "MinIO root secret key (MINIO_ROOT_PASSWORD)")
	serviceName     = flag.String("service-name", "ar-13-backend", "Name for the service account")
	userName        = flag.String("user-name", "ar-13-backend-user", "Name for the user")
)

func main() {
	flag.Parse()

	if *endpoint == "" {
		log.Fatal("Error: --endpoint is required (e.g., nas-ip:32774)")
	}
	if *accessKeyID == "" {
		log.Fatal("Error: --access-key is required (your MINIO_ROOT_USER)")
	}
	if *secretAccessKey == "" {
		log.Fatal("Error: --secret-key is required (your MINIO_ROOT_PASSWORD)")
	}

	fmt.Println("=== MinIO Access Key Creation ===")
	fmt.Printf("Endpoint: %s\n", *endpoint)
	fmt.Printf("Service Name: %s\n", *serviceName)
	fmt.Println()

	// Initialize MinIO admin client
	// Note: MinIO Go SDK doesn't have direct admin API support
	// We'll use the regular client and provide instructions for manual creation
	minioClient, err := minio.New(*endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(*accessKeyID, *secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("Error initializing MinIO client: %v", err)
	}

	ctx := context.Background()

	// Test connection
	fmt.Println("Testing connection...")
	buckets, err := minioClient.ListBuckets(ctx)
	if err != nil {
		log.Fatalf("Connection test failed: %v", err)
	}
	fmt.Printf("✓ Connected successfully! Found %d bucket(s)\n", len(buckets))
	fmt.Println()

	// Note: MinIO Go SDK doesn't support admin operations directly
	// We need to use the MinIO Client (mc) or Admin API
	fmt.Println("==========================================")
	fmt.Println("IMPORTANT: MinIO Go SDK doesn't support")
	fmt.Println("admin operations (creating access keys).")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("Please use one of these methods:")
	fmt.Println()
	fmt.Println("Method 1: Install MinIO Client (mc)")
	fmt.Println("  See: scripts/INSTALL_MINIO_CLIENT_WINDOWS.md")
	fmt.Println()
	fmt.Println("Method 2: Use MinIO Admin API directly")
	fmt.Println("  See instructions below:")
	fmt.Println()
	fmt.Println("Method 3: Use root credentials temporarily")
	fmt.Println("  MINIO_ACCESS_KEY=" + *accessKeyID)
	fmt.Println("  MINIO_SECRET_KEY=" + *secretAccessKey)
	fmt.Println("  (Not recommended for production)")
	fmt.Println()
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("To create access keys using MinIO Admin API:")
	fmt.Println()
	fmt.Println("1. Login to get session token:")
	fmt.Printf("   curl -X POST http://%s/api/v1/login \\\n", *endpoint)
	fmt.Println("     -H \"Content-Type: application/json\" \\")
	fmt.Printf("     -d '{\"accessKey\":\"%s\",\"secretKey\":\"%s\"}'\n", *accessKeyID, *secretAccessKey)
	fmt.Println()
	fmt.Println("2. Use the token to create service account")
	fmt.Println("   (See MINIO_SETUP_GUIDE.md for full instructions)")
}

