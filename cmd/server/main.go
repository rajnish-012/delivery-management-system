package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/rajnish-012/delivery-management-system/internal/api"
	"github.com/rajnish-012/delivery-management-system/internal/database"
)

func main() {
	ctx := context.Background()

	// Initialize PostgreSQL
	if err := database.InitPostgres(ctx); err != nil {
		log.Fatalf("postgres init failed: %v", err)
	}
	defer database.ClosePostgres()

	// Initialize Redis
	if err := database.InitRedis(ctx); err != nil {
		log.Fatalf("redis init failed: %v", err)
	}
	defer database.CloseRedis()

	//FIXED MIGRATION LOADING — PROPER IMPLEMENTATION
	migrationContent, err := os.ReadFile("migrations/0001_init.sql")
	if err != nil {
		log.Fatalf("failed to read migration file: %v", err)
	}
	if err := database.ExecMigration(ctx, string(migrationContent)); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	fmt.Println("Database initialized and migration applied successfully!")

	// Setup HTTP router
	r := mux.NewRouter()
	api.RegisterRoutes(r)

	// Setup server configuration
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in Goroutine
	go func() {
		fmt.Println("Server running on http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				fmt.Println("Server stopped gracefully")
			} else {
				log.Fatalf("Server failed: %v", err)
			}
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println("\nShutting down server...")
	ctxShut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShut); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	fmt.Println("Server exited properly")
}
