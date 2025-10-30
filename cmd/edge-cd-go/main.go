package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/config"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/files"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/git"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/pkgmgr"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/reconcile"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/svcmgr"
)

func main() {
	// Configure default slog handler (JSON handler for production)
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	slog.Info("Starting edge-cd-go")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	slog.Info("Configuration loaded successfully",
		"edgecd_repo", cfg.Spec.EdgeCD.Repo.URL,
		"config_repo", cfg.Spec.Config.Repo.URL,
		"polling_interval", cfg.Spec.PollingInterval,
	)

	// Wire dependencies: create all managers
	gitMgr := git.NewRepoManager()

	pkgMgr, err := pkgmgr.NewPackageManager(cfg.Spec.PackageManager.Name, cfg.EdgeCDRepoPath)
	if err != nil {
		slog.Error("Failed to create package manager", "error", err)
		os.Exit(1)
	}

	svcMgr, err := svcmgr.NewServiceManager(cfg.Spec.ServiceManager.Name, cfg.EdgeCDRepoPath)
	if err != nil {
		slog.Error("Failed to create service manager", "error", err)
		os.Exit(1)
	}

	fileRec := files.NewFileReconciler()

	// Create reconciler with all dependencies
	reconciler := reconcile.NewReconciler(cfg, gitMgr, pkgMgr, svcMgr, fileRec)

	// Set up context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start reconciler in a goroutine
	go func() {
		reconciler.Run(ctx)
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	slog.Info("Received shutdown signal", "signal", sig)

	// Trigger graceful shutdown
	cancel()

	slog.Info("edge-cd-go stopped")
	fmt.Println("edge-cd-go stopped successfully")
}
