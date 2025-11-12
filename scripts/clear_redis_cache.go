package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ar-13-go-backend/internal/config"
	"github.com/ar-13-go-backend/pkg/cache"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Redis connection
	client, err := cache.InitializeRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		fmt.Printf("❌ Redis connection failed: %v\n", err)
		fmt.Println("\nYour Redis configuration:")
		fmt.Printf("  Address: %s\n", cfg.RedisAddr)
		fmt.Printf("  Password: %s\n", func() string {
			if cfg.RedisPassword == "" {
				return "(empty)"
			}
			return "***"
		}())
		fmt.Printf("  Database: %d\n", cfg.RedisDB)
		os.Exit(1)
	}
	defer client.Close()

	// Create cache service
	cacheSvc := cache.NewCacheService()
	ctx := context.Background()

	// Get key count before clearing
	keys, err := client.Keys(ctx, "*").Result()
	if err != nil {
		log.Printf("Warning: Could not count keys: %v", err)
	} else {
		fmt.Printf("Found %d keys in cache\n", len(keys))
	}

	// Clear all cache
	fmt.Println("Clearing Redis cache...")
	if err := cacheSvc.FlushAll(ctx); err != nil {
		log.Fatalf("❌ Failed to clear cache: %v\n", err)
	}

	fmt.Println("✅ Redis cache cleared successfully!")
	fmt.Printf("  Database: %d\n", cfg.RedisDB)
	fmt.Printf("  Address: %s\n", cfg.RedisAddr)
}

