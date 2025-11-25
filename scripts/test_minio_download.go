package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	endpoint        = flag.String("endpoint", "", "MinIO/S3 endpoint (e.g., nas.example.com:9000) - overrides MINIO_ENDPOINT")
	accessKeyID     = flag.String("access-key", "", "Access Key ID - overrides MINIO_ACCESS_KEY")
	secretAccessKey = flag.String("secret-key", "", "Secret Access Key - overrides MINIO_SECRET_KEY")
	useSSL          = flag.Bool("use-ssl", false, "Use SSL/TLS connection - overrides MINIO_USE_SSL")
	insecureSSL     = flag.Bool("insecure-ssl", false, "Skip SSL certificate verification - overrides MINIO_INSECURE_SSL")
	bucketName      = flag.String("bucket", "", "Bucket name - overrides MINIO_BUCKET")
	objectName      = flag.String("object", "", "Object name to download (optional, if not provided lists all objects)")
	outputPath      = flag.String("output", "", "Output file path (defaults to object name in current directory)")
	listOnly        = flag.Bool("list", false, "Only list objects, don't download")
	prefix          = flag.String("prefix", "", "Filter objects by prefix")
	recursive       = flag.Bool("recursive", false, "List/download recursively (for folders)")
	region          = flag.String("region", "us-east-1", "Region (optional, some S3-compatible services ignore this)")
)

// getEnv reads environment variable with fallback to default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool reads boolean environment variable
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
		// Also accept "true"/"false" strings (case insensitive)
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

func main() {
	flag.Parse()

	// Read from environment variables (flags override env vars)
	endpointValue := getEnv("MINIO_ENDPOINT", "")
	if *endpoint != "" {
		endpointValue = *endpoint
	}

	accessKeyValue := getEnv("MINIO_ACCESS_KEY", "")
	if *accessKeyID != "" {
		accessKeyValue = *accessKeyID
	}

	secretKeyValue := getEnv("MINIO_SECRET_KEY", "")
	if *secretAccessKey != "" {
		secretKeyValue = *secretAccessKey
	}

	useSSLValue := getEnvBool("MINIO_USE_SSL", false)
	insecureSSLValue := getEnvBool("MINIO_INSECURE_SSL", false)

	// Check if flags were explicitly set by visiting all set flags
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "use-ssl" {
			useSSLValue = *useSSL
		}
		if f.Name == "insecure-ssl" {
			insecureSSLValue = *insecureSSL
		}
	})

	bucketValue := getEnv("MINIO_BUCKET", "ar-13-uploads")
	if *bucketName != "" {
		bucketValue = *bucketName
	}

	// Validate required values
	if endpointValue == "" {
		log.Fatal("Error: MINIO_ENDPOINT environment variable or --endpoint flag is required (e.g., nas.example.com:9000)")
	}
	if accessKeyValue == "" {
		log.Fatal("Error: MINIO_ACCESS_KEY environment variable or --access-key flag is required")
	}
	if secretKeyValue == "" {
		log.Fatal("Error: MINIO_SECRET_KEY environment variable or --secret-key flag is required")
	}

	// Update the flag values for use in the rest of the code
	*endpoint = endpointValue
	*accessKeyID = accessKeyValue
	*secretAccessKey = secretKeyValue
	*useSSL = useSSLValue
	*insecureSSL = insecureSSLValue
	*bucketName = bucketValue

	fmt.Println("=== MinIO/S3 Read/Download Test ===")
	fmt.Printf("Endpoint: %s\n", *endpoint)
	fmt.Printf("Use SSL: %v\n", *useSSL)
	fmt.Printf("Insecure SSL: %v\n", *insecureSSL)
	fmt.Printf("Bucket: %s\n", *bucketName)
	if *objectName != "" {
		fmt.Printf("Object: %s\n", *objectName)
	}
	if *prefix != "" {
		fmt.Printf("Prefix Filter: %s\n", *prefix)
	}
	fmt.Println()
	fmt.Println("Note: Values are read from environment variables (MINIO_*) or command-line flags")
	fmt.Println()

	// Initialize MinIO client
	minioClient, err := initializeMinIOClient()
	if err != nil {
		log.Fatalf("Error initializing MinIO client: %v", err)
	}

	ctx := context.Background()

	// Test 1: Check connection
	fmt.Println("Test 1: Testing connection...")
	if err := testConnection(ctx, minioClient); err != nil {
		log.Fatalf("Connection test failed: %v", err)
	}
	fmt.Println("✓ Connection successful!")
	fmt.Println()

	// Test 2: Check bucket exists
	fmt.Println("Test 2: Checking bucket...")
	if err := checkBucket(ctx, minioClient, *bucketName); err != nil {
		log.Fatalf("Bucket check failed: %v", err)
	}
	fmt.Printf("✓ Bucket '%s' exists!\n", *bucketName)
	fmt.Println()

	// Test 3: List or Download
	if *listOnly || *objectName == "" {
		fmt.Println("Test 3: Listing objects...")
		if err := listObjects(ctx, minioClient, *bucketName, *prefix, *recursive); err != nil {
			log.Fatalf("List objects failed: %v", err)
		}
	} else {
		fmt.Println("Test 3: Downloading object...")
		if err := downloadObject(ctx, minioClient, *bucketName, *objectName, *outputPath); err != nil {
			log.Fatalf("Download failed: %v", err)
		}
		fmt.Println("✓ Download successful!")
	}

	fmt.Println()
	fmt.Println("=== Operation completed! ===")
}

func initializeMinIOClient() (*minio.Client, error) {
	// Configure options
	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(*accessKeyID, *secretAccessKey, ""),
		Secure: *useSSL,
		Region: *region,
	}

	// Configure TLS if using SSL with insecure option
	if *useSSL && *insecureSSL {
		// For self-signed certificates, create custom HTTP transport
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		opts.Transport = tr
	}

	// Initialize MinIO client
	client, err := minio.New(*endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return client, nil
}

func testConnection(ctx context.Context, client *minio.Client) error {
	// Try to list buckets as a connection test
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("failed to list buckets: %w", err)
	}

	fmt.Printf("  Found %d bucket(s)\n", len(buckets))
	for _, bucket := range buckets {
		fmt.Printf("    - %s (created: %s)\n", bucket.Name, bucket.CreationDate.Format(time.RFC3339))
	}

	return nil
}

func checkBucket(ctx context.Context, client *minio.Client, bucketName string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Check if bucket exists
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("bucket '%s' does not exist", bucketName)
	}

	return nil
}

func listObjects(ctx context.Context, client *minio.Client, bucketName, prefix string, recursive bool) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	opts := minio.ListObjectsOptions{
		Recursive: recursive,
	}
	if prefix != "" {
		opts.Prefix = prefix
	}

	objectCh := client.ListObjects(ctx, bucketName, opts)

	count := 0
	totalSize := int64(0)

	fmt.Println("  Objects in bucket:")
	fmt.Println("  " + strings.Repeat("-", 80))
	fmt.Printf("  %-50s %15s %20s\n", "Name", "Size", "Last Modified")
	fmt.Println("  " + strings.Repeat("-", 80))

	for object := range objectCh {
		if object.Err != nil {
			return fmt.Errorf("error listing objects: %w", object.Err)
		}

		count++
		totalSize += object.Size

		// Format size
		sizeStr := formatSize(object.Size)
		// Format date
		dateStr := object.LastModified.Format("2006-01-02 15:04:05")

		fmt.Printf("  %-50s %15s %20s\n", object.Key, sizeStr, dateStr)
	}

	fmt.Println("  " + strings.Repeat("-", 80))
	fmt.Printf("  Total: %d object(s), %s\n", count, formatSize(totalSize))

	return nil
}

func downloadObject(ctx context.Context, client *minio.Client, bucketName, objectName, outputPath string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Get object info first
	objInfo, err := client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to stat object: %w", err)
	}

	fmt.Printf("  Object: %s\n", objectName)
	fmt.Printf("  Size: %s\n", formatSize(objInfo.Size))
	fmt.Printf("  Content Type: %s\n", objInfo.ContentType)
	fmt.Printf("  Last Modified: %s\n", objInfo.LastModified.Format(time.RFC3339))
	fmt.Printf("  ETag: %s\n", objInfo.ETag)
	fmt.Println()

	// Determine output path
	if outputPath == "" {
		outputPath = filepath.Base(objectName)
	}

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	if outputDir != "" && outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("  Warning: File '%s' already exists. Overwriting...\n", outputPath)
	}

	// Download the object
	fmt.Printf("  Downloading to: %s\n", outputPath)
	object, err := client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Copy object to file
	written, err := io.Copy(file, object)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if written != objInfo.Size {
		return fmt.Errorf("size mismatch: expected %d bytes, wrote %d bytes", objInfo.Size, written)
	}

	fmt.Printf("  ✓ Downloaded %s successfully\n", formatSize(written))

	return nil
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

