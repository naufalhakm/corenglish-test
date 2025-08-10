package worker

import (
	"context"
	"fmt"
	"go-corenglish/internal/config"
	"go-corenglish/pkg/database"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	logger *logrus.Logger
	redis  *redis.Client
}

func NewWorker(logger *logrus.Logger, redis *redis.Client) *Worker {
	return &Worker{
		logger: logger,
		redis:  redis,
	}
}

func (w *Worker) Start(ctx context.Context) {
	sub := w.redis.Subscribe(ctx, "tasks:invalidate")
	defer sub.Close()

	w.logger.Info("Worker subscribed to tasks:invalidate channel")

	ch := sub.Channel()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Worker shutting down...")
			return
		case msg := <-ch:
			w.handleMessage(ctx, msg.Payload)
		}
	}
}

func (w *Worker) handleMessage(ctx context.Context, userID string) {
	pattern := fmt.Sprintf("tasks:%s:*", userID)
	iter := w.redis.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()
		if err := w.redis.Del(ctx, key).Err(); err != nil {
			w.logger.WithError(err).Errorf("Failed to delete cache key %s", key)
		} else {
			w.logger.Infof("Deleted cache key: %s", key)
		}
	}
	if err := iter.Err(); err != nil {
		w.logger.WithError(err).Error("Error iterating Redis keys")
	}
}

func Run() {
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	logger := setupLogger(cfg)

	redisClient := database.ConnectRedis(cfg, logger)
	defer redisClient.Close()

	worker := NewWorker(logger, redisClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	worker.Start(ctx)
}

func setupLogger(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	if cfg.LogFormat == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// Set output to file in production
	if cfg.AppEnv == "production" {
		file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.Warn("Failed to open log file, using stdout")
		} else {
			logger.SetOutput(file)
		}
	}

	return logger
}
