package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	"example.com/taskservice/internal/jobs"
	postgresrepo "example.com/taskservice/internal/repository/postgres"
	transporthttp "example.com/taskservice/internal/transport/http"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
	"example.com/taskservice/internal/usecase/reminder"
	"example.com/taskservice/internal/usecase/task"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	taskRepo := postgresrepo.New(pool)
	taskUsecase := task.NewService(taskRepo)
	taskHandler := httphandlers.NewTaskHandler(taskUsecase)
	notificationHandler := httphandlers.NewNotificationHandler(taskRepo)
	docsHandler := swaggerdocs.NewHandler()
	router := transporthttp.NewRouter(taskHandler, notificationHandler, docsHandler)

	reminderSvc := reminder.NewService(taskRepo, taskRepo)
	if cfg.ReminderEnabled {
		go jobs.RunMorningReminderLoop(ctx, logger, reminderSvc, cfg.ReminderAtUTC)
		logger.Info("morning reminders enabled", "at_utc", cfg.ReminderAtUTC)
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("http server started", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen and serve", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown http server", "error", err)
	}
}

type config struct {
	HTTPAddr        string
	DatabaseDSN     string
	ReminderEnabled bool
	ReminderAtUTC   string
}

func loadConfig() config {
	cfg := config{
		HTTPAddr:        envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseDSN:     envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"),
		ReminderEnabled: envOrDefault("REMINDER_ENABLED", "true") != "false",
		ReminderAtUTC:   envOrDefault("REMINDER_AT_UTC", "08:00"),
	}

	if cfg.DatabaseDSN == "" {
		panic(fmt.Errorf("DATABASE_DSN is required"))
	}

	return cfg
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
