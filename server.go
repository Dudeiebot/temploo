package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Dudeiebot/dlog"
	"github.com/gorilla/mux"
	"github.com/hibiken/asynq"
	"github.com/hibiken/asynqmon"

	"github.com/dudeiebot/ad-ly/config"
	"github.com/dudeiebot/ad-ly/queue"
	"github.com/dudeiebot/ad-ly/routes"
)

/*
we have three main things that need to run simultaneously:

The main server
The monitoring server
The Asynq worker

Each of these needs to keep running and listening for requests/jobs continuously.

We add 1 to the counter for each goroutine we start
Each goroutine calls wg.Done() when it finishes
We use wg.Wait() to wait until all goroutines are done

This ensures we don't exit the program while servers are still shutting down.
*/

var logger = dlog.NewLog(dlog.LevelTrace)

func Init() {
	// Initialize error channel and signal handling
	serverError := make(chan error, 3)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Initialize components with proper error handling
	if err := initialize(); err != nil {
		logger.Info("Intialization Error", err)
		return
	}

	// Start servers
	var wg sync.WaitGroup
	_, cancel := context.WithCancel(context.Background())

	// Start Cook server
	server := startServer(&wg, serverError)

	// Start monitoring server conditionally
	var monitoringServer *http.Server
	if config.AppConfig.AsynqmonService == "true" {
		monitoringServer = startMonitoringServer(&wg, serverError)
	}
	// Start Asynq worker
	worker := startAsynqWorker(&wg, serverError)

	// Wait for shutdown signal or error
	shutdown := false
	select {
	case <-stop:
		fmt.Println("Shutdown signal received...")
		shutdown = true
	case err := <-serverError:
		logger.Info("Server error triggered shutdown", err)
	}

	// Cancel context to signal all operations to stop
	cancel()

	// Start graceful shutdown
	fmt.Println("Initiating graceful shutdown...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown servers
	shutdownServer(shutdownCtx, server, "Http Server")
	if monitoringServer != nil {
		shutdownServer(shutdownCtx, monitoringServer, "Monitoring server")
	}
	if worker != nil {
		shutdownWorker(shutdownCtx, worker)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close database connections
	closeConnections()

	if shutdown {
		fmt.Println("All servers have been stopped gracefully.")
	} else {
		fmt.Println("Servers stopped due to error.")
	}
}

// initialize sets up all required dependencies
func initialize() error {
	err := config.LoadEnvironmentVariable()
	if err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	err = config.ConnectPostGres(&config.DbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to PostGre: %w", err)
	}

	err = config.ConnectRedis(&config.DbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// startCookServer initializes and starts the main application server
func startServer(wg *sync.WaitGroup, serverError chan<- error) *http.Server {
	server := &http.Server{
		Addr:    ":6060",
		Handler: routes.Routes(),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Starting Cook server on :6060")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverError <- fmt.Errorf("cook server error: %v", err)
		}
	}()

	return server
}

// startMonitoringServer initializes and starts the monitoring server
func startMonitoringServer(wg *sync.WaitGroup, serverError chan<- error) *http.Server {
	m := asynqmon.New(asynqmon.Options{
		RootPath: "/monitoring",
		RedisConnOpt: asynq.RedisClientOpt{
			Addr: ":" + config.DbConfig.RedisPort,
		},
	})

	router := mux.NewRouter()
	router.PathPrefix(m.RootPath()).Handler(m)

	server := &http.Server{
		Handler: m,
		Addr:    ":6660",
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Starting Monitoring server on :6660")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverError <- fmt.Errorf("monitoring server error: %v", err)
		}
	}()
	fmt.Printf("Monitoring server setup done. Visit http://localhost:6660%v\n", m.RootPath())
	return server
}

// startAsynqWorker initializes and starts the Asynq worker
func startAsynqWorker(wg *sync.WaitGroup, serverError chan<- error) *asynq.Server {
	worker := asynq.NewServer(
		asynq.RedisClientOpt{Addr: config.Redis.Options().Addr},
		asynq.Config{
			Concurrency:     10,
			ShutdownTimeout: 8 * time.Second, // Allow time for graceful shutdown
		},
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Starting Asynq worker server")
		if err := worker.Run(queue.Register()); err != nil {
			serverError <- fmt.Errorf("asynq worker error: %v", err)
		}
	}()

	return worker
}

// shutdownServer gracefully shuts down an HTTP server
func shutdownServer(ctx context.Context, server *http.Server, name string) {
	if server == nil {
		return
	}

	fmt.Printf("Shutting down %s...\n", name)
	if err := server.Shutdown(ctx); err != nil {
		logger.Info("shutdonw failed", err)
	}
}

// shutdownWorker gracefully shuts down the Asynq worker
func shutdownWorker(ctx context.Context, worker *asynq.Server) {
	if worker == nil {
		return
	}

	fmt.Println("Shutting down Asynq worker...")
	worker.Shutdown()
}

// closeConnections closes all database connections
func closeConnections() {
	if err := queue.Client.Close(); err != nil {
		logger.Info("Failed to close redis queue", err)
	}

	if err := config.Redis.Close(); err != nil {
		logger.Info("Failed to close Redis Connection", err)
	}

	db, err := config.PostDb.DB()
	if err != nil {
		logger.Info("Failed to get db from gorm", err)
	} else {
		if err := db.Close(); err != nil {
			logger.Info("Failed to close PostGres Db", err)
		}
	}
}
