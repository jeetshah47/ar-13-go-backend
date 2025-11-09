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

	// Try to connect to Redis
	client, err := cache.InitializeRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		fmt.Printf("❌ Redis connection failed: %v\n", err)
		fmt.Println("\nYour Redis configuration:")
		fmt.Printf("  Address: %s\n", cfg.RedisAddr)
		fmt.Printf("  Password: %s (empty = no password)\n", func() string {
			if cfg.RedisPassword == "" {
				return "(empty)"
			}
			return "***"
		}())
		fmt.Printf("  Database: %d\n", cfg.RedisDB)
		os.Exit(1)
	}

	// Test connection
	ctx := context.Background()
	cacheSvc := cache.NewCacheService()
	
	err = cacheSvc.Set(ctx, "test", "hello", 0)
	if err != nil {
		fmt.Printf("❌ Redis write test failed: %v\n", err)
		os.Exit(1)
	}

	var result string
	err = cacheSvc.Get(ctx, "test", &result)
	if err != nil {
		fmt.Printf("❌ Redis read test failed: %v\n", err)
		os.Exit(1)
	}

	// Cleanup
	_ = cacheSvc.Delete(ctx, "test")
	_ = client.Close()

	fmt.Println("✅ Redis connection successful!")
	fmt.Println("\nYour Redis configuration:")
	fmt.Printf("  Address: %s\n", cfg.RedisAddr)
	fmt.Printf("  Password: %s\n", func() string {
		if cfg.RedisPassword == "" {
			return "(none - no password required)"
		}
		return "*** (password is set)"
	}())
	fmt.Printf("  Database: %d\n", cfg.RedisDB)
}

