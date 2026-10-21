package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	routes "padi-template/app/Routes"
	_ "padi-template/database/migrations"
	"github.com/wibiesana/padi_go_core/config"
	"github.com/wibiesana/padi_go_core/database"
	"github.com/wibiesana/padi_go_core/migrator"
	"github.com/wibiesana/padi_go_core/router"
)

func main() {
	cfg := config.Load()

	// Connect to Database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("❌ Database connection error: %v", err)
	}
	log.Printf("🔌 Connected to database [%s]", cfg.DBConnection)

	// Run pending migrations in development automatically
	if cfg.AppEnv == "development" {
		if err := migrator.RunPending(db); err != nil {
			log.Printf("⚠️ Migration notice: %v", err)
		}
	}

	// Initialize Router & Register Routes
	r := router.New(cfg)
	routes.RegisterRoutes(r)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("🌾 %s running on http://localhost%s (Env: %s)", cfg.AppName, addr, cfg.AppEnv)

	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown: trap SIGINT / SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Block until a signal is received
	<-ctx.Done()
	stop() // release signal resources immediately

	log.Println("🛑 Shutdown signal received, gracefully shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}
	log.Println("✅ Server shutdown complete")
}
