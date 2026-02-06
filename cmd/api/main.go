// Package main is the entry point for the API server.
package main

import (
	"fmt"
	"log"
	"net/http"

	"api-test/internal/cache"
	"api-test/internal/config"
	"api-test/internal/database"
	"api-test/internal/handler"
	"api-test/internal/service"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db := database.Connect(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	defer db.Close()

	// Initialize Redis cache
	redisCache := cache.New(cfg.RedisHost, cfg.RedisPort)
	defer redisCache.Close()

	// Initialize service layer
	userService := service.New(db, redisCache)

	// Initialize handlers
	userHandler := handler.New(userService)

	// Register routes
	handler.RegisterRoutes(userHandler)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("Server running on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
