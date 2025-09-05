package main

import (
	"blackbird-eu/waypoints2-v2/config"
	"blackbird-eu/waypoints2-v2/internal/rabbitmq"
	"blackbird-eu/waypoints2-v2/pkg/logger"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	logger.Log.Out = os.Stdout
}

func main() {
	// Load config
	config.LoadConfig()

	// Initialize RabbitMQ instance
	rmq, err := rabbitmq.NewRabbitMQ(&config.AppConfig.RabbitMQ)
	if err != nil {
		logger.Log.Errorf("Failed to initialize RabbitMQ: %v", err)
	}
	defer rmq.Close()

	logger.Log.Info("🟦 Consumer Starting...")
	err = rmq.StartConsumer()
	if err != nil {
		logger.Log.Errorf("Failed to consume message: %v", err)
	}

	logger.Log.Info("✅ Service started...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			count := 0
			rmq.ProcessingTaskMap.Range(func(_, _ interface{}) bool {
				count++
				return true // Continue iteration
			})

			if count > 0 {
				logger.Log.Infof("There are existing tasks: %d", count)
			} else {
				// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				// defer cancel()

				// if err := srv.Shutdown(ctx); err != nil {
				// 	logger.Log.Fatal("⚠️ Server Shutdown Error: ", err)
				// }
				logger.Log.Info("✅ Service was closed successfully")
				return
			}
		}
	}
}
