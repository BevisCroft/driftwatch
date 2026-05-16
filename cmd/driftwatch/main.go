// Command driftwatch is the main entry point for the driftwatch daemon.
// It wires together configuration loading, scheduling, alerting, health checks,
// and metrics exposure into a single long-running process.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/driftwatch/internal/alerting"
	"github.com/yourusername/driftwatch/internal/config"
	"github.com/yourusername/driftwatch/internal/healthcheck"
	"github.com/yourusername/driftwatch/internal/metrics"
	"github.com/yourusername/driftwatch/internal/reporter"
	"github.com/yourusername/driftwatch/internal/scheduler"
	"github.com/yourusername/driftwatch/internal/snapshot"
)

func main() {
	cfgPath := flag.String("config", "", "path to driftwatch config file (optional)")
	flag.Parse()

	// Initialise structured logger.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration, falling back to defaults when no file is provided.
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}
	slog.Info("configuration loaded",
		"manifest_dir", cfg.ManifestDir,
		"poll_interval", cfg.PollInterval,
		"alert_level", cfg.AlertLevel,
	)

	// Build shared subsystems.
	snapshotStore := snapshot.NewStore(cfg.SnapshotDir)
	alert := alerting.New(cfg.AlertLevel, logger)
	rep := reporter.New(cfg.OutputFormat, os.Stdout)
	met := metrics.New()
	hc := healthcheck.New()

	// Create the drift scheduler that ties everything together.
	sched := scheduler.New(scheduler.Options{
		ManifestDir:   cfg.ManifestDir,
		PollInterval:  time.Duration(cfg.PollInterval),
		Store:         snapshotStore,
		Alerter:       alert,
		Reporter:      rep,
		Metrics:       met,
		Healthcheck:   hc,
		Logger:        logger,
	})

	// Set up the HTTP server for health and metrics endpoints.
	mux := http.NewServeMux()
	mux.Handle("/healthz", hc.Handler())
	mux.Handle("/metrics", met.Handler())

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Handle OS signals for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start HTTP server in the background.
	go func() {
		slog.Info("HTTP server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	// Run the scheduler until the context is cancelled.
	slog.Info("driftwatch daemon starting")
	if err := sched.Run(ctx); err != nil {
		slog.Error("scheduler exited with error", "error", err)
	}

	// Gracefully shut down the HTTP server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	}

	slog.Info("driftwatch daemon stopped")
}
