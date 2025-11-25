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
	filePath        = flag.String("file", "", "Path to file to upload")
	objectName      = flag.String("object", "", "Object name in bucket (defaults to filename)")
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
	if *filePath == "" {
		log.Fatal("Error: --file is required (path to file to upload)")
	}

	// Update the flag values for use in the rest of the code
	*endpoint = endpointValue
	*accessKeyID = accessKeyValue
	*secretAccessKey = secretKeyValue
	*useSSL = useSSLValue
	*insecureSSL = insecureSSLValue
	*bucketName = bucketValue

	// Check if file exists
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		log.Fatalf("Error: File does not exist: %s", *filePath)
	}

	// Set default object name if not provided
	if *objectName == "" {
		*objectName = filepath.Base(*filePath)
	}

	fmt.Println("=== MinIO/S3 Connection Test ===")
	fmt.Printf("Endpoint: %s\n", *endpoint)
	fmt.Printf("Use SSL: %v\n", *useSSL)
	fmt.Printf("Insecure SSL: %v\n", *insecureSSL)
	fmt.Printf("Bucket: %s\n", *bucketName)
	fmt.Printf("File: %s\n", *filePath)
	fmt.Printf("Object Name: %s\n", *objectName)
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

	// Test 2: Check/Create bucket
	fmt.Println("Test 2: Checking bucket...")
	if err := ensureBucket(ctx, minioClient, *bucketName); err != nil {
		log.Fatalf("Bucket check/create failed: %v", err)
	}
	fmt.Printf("✓ Bucket '%s' is ready!\n", *bucketName)
	fmt.Println()

	// Test 3: Upload file
	fmt.Println("Test 3: Uploading file...")
	fileInfo, err := uploadFile(ctx, minioClient, *bucketName, *filePath, *objectName)
	if err != nil {
		log.Fatalf("File upload failed: %v", err)
	}
	fmt.Printf("✓ File uploaded successfully!\n")
	fmt.Printf("  Object: %s\n", *objectName)
	fmt.Printf("  Size: %d bytes\n", fileInfo.Size)
	fmt.Printf("  ETag: %s\n", fileInfo.ETag)
	fmt.Println()

	// Test 4: Verify upload (list objects)
	fmt.Println("Test 4: Verifying upload...")
	if err := verifyUpload(ctx, minioClient, *bucketName, *objectName); err != nil {
		log.Fatalf("Upload verification failed: %v", err)
	}
	fmt.Printf("✓ Upload verified! Object '%s' exists in bucket '%s'\n", *objectName, *bucketName)
	fmt.Println()

	fmt.Println("=== All tests passed! ===")
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
		// Provide more helpful error messages
		errMsg := err.Error()
		if strings.Contains(errMsg, "Access Denied") {
			return fmt.Errorf("access denied - check if your access key has 'ListAllMyBuckets' permission. Error: %w", err)
		}
		if strings.Contains(errMsg, "time") && strings.Contains(errMsg, "too large") {
			return fmt.Errorf("time sync error - sync your system clock. Error: %w", err)
		}
		return fmt.Errorf("failed to list buckets: %w", err)
	}

	fmt.Printf("  Found %d bucket(s)\n", len(buckets))
	for _, bucket := range buckets {
		fmt.Printf("    - %s (created: %s)\n", bucket.Name, bucket.CreationDate.Format(time.RFC3339))
	}

	return nil
}

func ensureBucket(ctx context.Context, client *minio.Client, bucketName string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Check if bucket exists
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		fmt.Printf("  Bucket '%s' does not exist. Creating...\n", bucketName)
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
			Region: *region,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		fmt.Printf("  ✓ Bucket '%s' created successfully\n", bucketName)
	} else {
		fmt.Printf("  ✓ Bucket '%s' already exists\n", bucketName)
	}

	return nil
}

func uploadFile(ctx context.Context, client *minio.Client, bucketName, filePath, objectName string) (*minio.UploadInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileStat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	// Detect content type
	contentType := "application/octet-stream"
	ext := filepath.Ext(filePath)
	switch ext {
	case ".txt":
		contentType = "text/plain"
	case ".json":
		contentType = "application/json"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".pdf":
		contentType = "application/pdf"
	case ".zip":
		contentType = "application/zip"
	}

	fmt.Printf("  Uploading: %s (%d bytes)\n", filePath, fileStat.Size())
	fmt.Printf("  Content Type: %s\n", contentType)

	// Upload the file
	uploadInfo, err := client.PutObject(ctx, bucketName, objectName, file, fileStat.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &uploadInfo, nil
}

func verifyUpload(ctx context.Context, client *minio.Client, bucketName, objectName string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get object info
	objInfo, err := client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to stat object: %w", err)
	}

	fmt.Printf("  Object found: %s\n", objectName)
	fmt.Printf("  Size: %d bytes\n", objInfo.Size)
	fmt.Printf("  Last Modified: %s\n", objInfo.LastModified.Format(time.RFC3339))
	fmt.Printf("  ETag: %s\n", objInfo.ETag)

	// Optionally, download and verify a small portion
	fmt.Println("  Verifying file integrity...")
	reader, err := client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer reader.Close()

	// Read first 100 bytes to verify
	buffer := make([]byte, 100)
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read object: %w", err)
	}
	fmt.Printf("  ✓ Successfully read %d bytes from uploaded file\n", n)

	return nil
}
