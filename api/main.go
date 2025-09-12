package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.json", "Path to GhostFS configuration file")
	flag.Parse()

	// Create GhostFS server
	server, err := NewGhostFSServer(configPath)
	if err != nil {
		log.Fatalf("Failed to create GhostFS server: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("🛑 Shutdown signal received, stopping server...")
		cancel()
	}()

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	// Wait for shutdown
	<-ctx.Done()

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Stop(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	} else {
		log.Println("✅ Server stopped gracefully")
	}
}
