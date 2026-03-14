package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/pidrive/pidrive/api"
	"github.com/pidrive/pidrive/internal/auth"
	"github.com/pidrive/pidrive/internal/config"
	"github.com/pidrive/pidrive/internal/db"
	"github.com/pidrive/pidrive/internal/sftpd"
)

func main() {
	cfg := config.Load()

	log.Println("Connecting to database...")
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	log.Println("Running migrations...")
	execPath, _ := os.Executable()
	migrationsDir := filepath.Join(filepath.Dir(execPath), "..", "..", "migrations")
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = "migrations"
	}
	if err := database.Migrate(migrationsDir); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	server := api.NewServer(cfg, database)

	// Start search indexer
	log.Println("Starting search indexer...")
	server.StartIndexer()
	defer server.StopIndexer()

	// Start SFTP server
	authSvc := auth.NewAuthService(database)
	sftpAddr := fmt.Sprintf(":%s", cfg.SFTPPort)
	sftpServer := sftpd.NewSFTPServer(authSvc, cfg.JuiceFSMountPath, cfg.HostKeyPath, sftpAddr)
	if err := sftpServer.Start(); err != nil {
		log.Fatalf("Failed to start SFTP server: %v", err)
	}

	router := server.Router()

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("pidrive server starting on %s", addr)
	log.Printf("  Server URL: %s", cfg.ServerURL)
	log.Printf("  SFTP:       %s", sftpAddr)
	log.Printf("  Database:   %s", maskURL(cfg.DatabaseURL))
	log.Printf("  Redis:      %s", cfg.RedisURL)
	log.Printf("  Mount:      %s", cfg.JuiceFSMountPath)
	log.Printf("  OS/Arch:    %s/%s", runtime.GOOS, runtime.GOARCH)

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := http.ListenAndServe(addr, router); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down...")
}

func maskURL(url string) string {
	if len(url) > 20 {
		return url[:20] + "..."
	}
	return url
}
