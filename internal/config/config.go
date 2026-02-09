package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server
	Port    int
	NodeEnv string

	// MongoDB
	MongoDBURI      string
	MongoDBDatabase string

	// JWT
	JWTSecret         string
	JWTExpiration     int // in hours
	RefreshExpiration int // in days

	// Email
	EmailHost     string
	EmailPort     int
	EmailUser     string
	EmailPassword string
	EmailFrom     string
	EmailFromName string

	// Frontend
	FrontendURL string

	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// MinIO/NAS Storage
	MinIOEndpoint    string
	MinIOAccessKey   string
	MinIOSecretKey   string
	MinIOBucket      string
	MinIOUseSSL      bool
	MinIOInsecureSSL bool

	// FileBrowser Storage (alternative to MinIO)
	// Note: FileBrowserEnabled and FileBrowserToken are for the old FileBrowser app
	// FileBrowserServiceURL is for the new simple Go filebrowser service
	FileBrowserEnabled          bool
	FileBrowserURL              string
	FileBrowserToken      string
	FileBrowserServiceURL string // URL for the new filebrowser service (e.g., http://localhost:8082)
	// Note: FileBrowserServiceSecretKey removed - now using JWT authentication

	// QNAP NAS Integration
	QNAPNASIP           string // QNAP NAS IP address (e.g., 192.168.1.100)
	QNAPAPIPort         int    // QNAP API port (default: 8080)
	QNAPServiceUser     string // Service account username
	QNAPServicePassword string // Service account password
	QNAPShareName       string // Share name (e.g., "studio-work")
	QNAPSessionTimeout  int    // Session timeout in minutes (default: 30)
	NASBasePath         string // Base path for direct filesystem access (when backend is on NAS, e.g., "/share/studio-work")
}

var AppConfig *Config

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("PORT", "3000"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	emailPort, err := strconv.Atoi(getEnv("EMAIL_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("invalid EMAIL_PORT: %w", err)
	}

	config := &Config{
		Port:    port,
		NodeEnv: getEnv("NODE_ENV", "development"),

		MongoDBURI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBDatabase: getEnv("MONGODB_DATABASE", "ar13_backend"),

		JWTSecret:         strings.TrimSpace(getEnv("JWT_SECRET", "")),
		JWTExpiration:     getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		RefreshExpiration: getEnvAsInt("REFRESH_EXPIRATION_DAYS", 30),

		EmailHost:     getEnv("EMAIL_HOST", ""),
		EmailPort:     emailPort,
		EmailUser:     getEnv("EMAIL_USER", ""),
		EmailPassword: getEnv("EMAIL_PASSWORD", ""),
		EmailFrom:     getEnv("EMAIL_FROM", ""),
		EmailFromName: getEnv("EMAIL_FROM_NAME", ""),

		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		MinIOEndpoint:    getEnv("MINIO_ENDPOINT", ""),
		MinIOAccessKey:   getEnv("MINIO_ACCESS_KEY", ""),
		MinIOSecretKey:   getEnv("MINIO_SECRET_KEY", ""),
		MinIOBucket:      getEnv("MINIO_BUCKET", "ar-13-uploads"),
		MinIOUseSSL:      getEnv("MINIO_USE_SSL", "false") == "true",
		MinIOInsecureSSL: getEnv("MINIO_INSECURE_SSL", "false") == "true",

		FileBrowserEnabled:    getEnv("FILEBROWSER_ENABLED", "false") == "true",
		FileBrowserURL:        getEnv("FILEBROWSER_URL", "http://localhost:8080"),
		FileBrowserToken:      getEnv("FILEBROWSER_TOKEN", ""),
		FileBrowserServiceURL: getEnv("FILEBROWSER_SERVICE_URL", "https://api.jsdeveloper.cloud"),
		// Note: FileBrowserServiceSecretKey removed - now using JWT authentication

		QNAPNASIP:           getEnv("QNAP_NAS_IP", ""),
		QNAPAPIPort:         getEnvAsInt("QNAP_API_PORT", 8080),
		QNAPServiceUser:     getEnv("QNAP_SERVICE_USER", ""),
		QNAPServicePassword: getEnv("QNAP_SERVICE_PASSWORD", ""),
		QNAPShareName:       getEnv("QNAP_SHARE_NAME", "studio-work"),
		QNAPSessionTimeout:  getEnvAsInt("QNAP_SESSION_TIMEOUT", 30),
		NASBasePath:         getEnv("NAS_BASE_PATH", ""),
	}

	AppConfig = config
	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.NodeEnv == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.NodeEnv == "production"
}
