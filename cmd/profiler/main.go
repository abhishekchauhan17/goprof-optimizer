package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AbhishekChauhan17/goprof-optimizer/internal/alerts"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/config"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/health"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/logging"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/metrics"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/profiler"
	"github.com/AbhishekChauhan17/goprof-optimizer/internal/version"
)

func main() {
	// ---- CLI Flags ----
	var cfgPath string
	var showVersion bool

	flag.StringVar(&cfgPath, "config", "", "Path to configuration file (YAML or JSON)")
	flag.BoolVar(&showVersion, "version", false, "Print version information and exit")
	flag.Parse()

	if showVersion {
		fmt.Println("goprof-optimizer", version.String())
		return
	}

	// ---- Load config ----
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// ---- Logger ----
	logger := logging.NewLogger(cfg.LogLevel)
	logger.Info("starting goprof-optimizer", "version", version.String())

	// ---- Profiler ----
	prof := profiler.NewProfiler(cfg, logger)
	alertEngine := alerts.NewEngine()
	healthChecker := health.NewChecker(cfg, prof)

	// ---- HTTP Server ----
	srv := metrics.NewServer(cfg, prof, alertEngine, healthChecker, logger)
	router := srv.Router()

	httpServer := &http.Server{
		Addr:         cfg.MetricsListenAddr,
		Handler:      router,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// ---- Optional separate Pprof server ----
	var pprofServer *http.Server
	if cfg.PprofEnabled && cfg.PprofListenAddr != "" {
		pprofMux := http.NewServeMux()
		profiler.RegisterPprofHandlers(pprofMux)
		pprofServer = &http.Server{
			Addr:    cfg.PprofListenAddr,
			Handler: pprofMux,
		}
	}

	// ---- Context for graceful shutdown ----
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ---- Start profiler sampling ----
	logger.Info("starting profiler sampling")
	prof.Start(ctx)

	// ---- Start HTTP server ----
	go func() {
		logger.Info("HTTP server starting", "addr", cfg.MetricsListenAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server crashed", "error", err.Error())
			cancel()
		}
	}()

	// ---- Start separate pprof server if configured ----
	if pprofServer != nil {
		go func() {
			logger.Info("pprof server starting", "addr", cfg.PprofListenAddr)
			if err := pprofServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("pprof server crashed", "error", err.Error())
				cancel()
			}
		}()
	}

	// ---- OS Signal Handling ----
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("received shutdown signal", "signal", sig.String())
	case <-ctx.Done():
		logger.Warn("context cancelled, shutting down")
	}

	// ---- Graceful Shutdown ----
	shutdownTimeout := time.Duration(cfg.ShutdownGracePeriodSec) * time.Second
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	logger.Info("shutting down HTTP server gracefully", "deadline_sec", cfg.ShutdownGracePeriodSec)
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP shutdown failed", "error", err.Error())
	}

	if pprofServer != nil {
		logger.Info("shutting down pprof server")
		if err := pprofServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("pprof shutdown failed", "error", err.Error())
		}
	}

	logger.Info("shutdown complete")
}
