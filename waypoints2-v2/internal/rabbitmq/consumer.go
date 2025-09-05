package rabbitmq

import (
	"blackbird-eu/waypoints2-v2/pkg/logger"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type ScanInfo struct {
	// Required Fields
	ScanID         uuid.UUID `json:"scanID"`
	CustomerAPIKey string    `json:"customerAPIKey"`
	CustomerID     uuid.UUID `json:"customerID"`
	Target         string    `json:"target"`
	TemplateIDs    []string  `json:"templateIDs"`

	// Optional Fields
	TargetID                 uuid.UUID `json:"targetID"`
	VulnerabilityScanID      uuid.UUID `json:"vulnerabilityScanID"`
	SubDomainxScanID         uuid.UUID `json:"subDomainxScanID"`
	SpiderxScanID            uuid.UUID `json:"spiderxScanID"`
	ChunkID                  uint64    `json:"chunkID"`
	TemplateTags             string    `json:"templateTags"`
	Headers                  []string  `json:"headers"`
	Delay                    uint64    `json:"delay"`
	Timeout                  uint64    `json:"timout"`
	VpnConnectionURI         string    `json:"vpnConnectionURI"`
	CustomerNotificationsKey string    `json:"customerNotificationsKey"`
	CustomerCanaryToken      string    `json:"customerCanaryToken"`
}

type Task struct {
	ID          string
	ScanInfo    ScanInfo
	StartedAt   time.Time
	Completion  time.Time
	IsCompleted bool
}

type RabbitMQ struct {
	Config *Config

	Conn    *amqp.Connection
	Channel *amqp.Channel

	ProcessingTaskMap sync.Map
	stopChan          chan os.Signal
}

func NewRabbitMQ(cfg *Config) (*RabbitMQ, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %v", err)
	}

	// Declare the queue
	_, err = ch.QueueDeclare(
		cfg.QueueName, // Queue name
		cfg.Durable,   // Durable
		false,         // Delete when unused
		false,         // Exclusive
		false,         // No-wait
		nil,           // Arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	return &RabbitMQ{Conn: conn, Channel: ch, Config: cfg, stopChan: sigChan}, nil
}

// Producer sends a message to the queue.
func (r *RabbitMQ) Producer(message string) error {
	err := r.Channel.Publish(
		"",                 // Exchange (empty means default)
		r.Config.QueueName, // Routing key (queue name)
		false,              // Mandatory
		false,              // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}
	fmt.Printf("Sent: %s\n", message)
	return nil
}

// Consumer listens for messages and processes them using the provided handler function.
func (r *RabbitMQ) StartConsumer() error {
	msgs, err := r.Channel.Consume(
		r.Config.QueueName, // Queue name
		"",                 // Consumer tag (empty means default)
		r.Config.AutoAck,   // Auto-acknowledge
		false,              // Exclusive
		false,              // No-local
		false,              // No-wait
		nil,                // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %v", err)
	}

	go r.handleConsumerMsg(msgs)
	return nil
}

func (r *RabbitMQ) handleConsumerMsg(msgs <-chan amqp.Delivery) {

	for {
		select {
		case msg := <-msgs:
			taskID := fmt.Sprintf("%d", msg.DeliveryTag)

			var scanInfo ScanInfo
			err := json.Unmarshal(msg.Body, &scanInfo)
			if err != nil {
				logger.Log.Error("Error parsing order message JSON: ", err)
				continue
			}

			task := &Task{
				ID:        taskID,
				ScanInfo:  scanInfo,
				StartedAt: time.Now(),
			}

			// Store task in the map to track its progress
			r.ProcessingTaskMap.Store(taskID, task)

			// Start a new goroutine to process the message
			go r.processTask(task)

			// Acknowledge the message after it has been processed
			// Acknowledging before starting processing allows to have it marked as handled
			if !r.Config.AutoAck {
				r.Channel.Ack(msg.DeliveryTag, false)
			}
		case <-r.stopChan:
			logger.Log.Info("✅ Stoped the consumer")
			signal.Stop(r.stopChan)
			close(r.stopChan)
			return
		}
	}
}

// processTask simulates processing of a long-running task.
func (r *RabbitMQ) processTask(task *Task) {
	// Simulate processing time (at least 2 hours)
	logger.Log.Debugf("Processing task: %s, ScanID: %s, CustomerID: %s\n", task.ID, task.ScanInfo.ScanID, task.ScanInfo.CustomerID)

	// TODO Implement the scanner

	// Remove the task from the map once processing is done
	r.ProcessingTaskMap.Delete(task.ID)
}

// Close closes the connection and channel gracefully.
func (r *RabbitMQ) Close() {
	if err := r.Channel.Close(); err != nil {
		log.Printf("Failed to close channel: %v", err)
	}
	if err := r.Conn.Close(); err != nil {
		log.Printf("Failed to close connection: %v", err)
	}
}
