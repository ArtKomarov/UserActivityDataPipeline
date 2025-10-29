package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
)

type Event struct {
	UserID    string `json:"user_id"`
	EventType string `json:"event_type"`
	Timestamp int64  `json:"timestamp"`
	URL       string `json:"url"`
}

const (
	topic         = "user_events"
	brokerAddress = "kafka:19092"
)

var (
	eventsProduced = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_producer_events_total",
			Help: "Total number of user events produced to Kafka.",
		},
	)
)

func main() {
	if err := createTopicIfNotExists(brokerAddress, topic); err != nil {
		log.Fatalf("Fatal: Could not initialize Kafka topic: %v", err)
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:   []string{brokerAddress},
		Topic:     topic,
		BatchSize: 1,
		Balancer:  &kafka.LeastBytes{},
	})
	defer writer.Close()

	log.Println("Starting metrics server on :8080")
	go setupMetrics()

	log.Println("Kafka Producer started. Generating events...")

	for {
		event := generateEvent()
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Error marshaling event: %v", err)
			continue
		}

		msg := kafka.Message{
			Value: eventJSON,
		}

		err = writer.WriteMessages(context.Background(), msg)
		if err != nil {
			log.Printf("Failed to write message to Kafka: %v", err)
		} else {
			eventsProduced.Inc()
			log.Printf("Produced event: UserID=%s, Type=%s, URL=%s", event.UserID, event.EventType, event.URL)
		}

		sleepDuration := time.Duration(rand.IntN(4)*500+1000) * time.Millisecond
		time.Sleep(sleepDuration)
	}
}

func setupMetrics() {
	prometheus.MustRegister(eventsProduced)
	http.Handle("/metrics", promhttp.Handler())
	// Run on the default port 8080 inside the container
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// createTopicIfNotExists creates a Kafka topic if it does not already exist
func createTopicIfNotExists(brokerAddress, topicName string) error {
	retriesNumber := 5
	retryInterval := 3 * time.Second
	var conn *kafka.Conn
	var err error

	log.Println("Initializing Kafka connection...")
	for i := range retriesNumber {
		log.Printf("Attempting to connect to Kafka (Attempt %d/%d)...", i+1, retriesNumber)
		conn, err = kafka.Dial("tcp", brokerAddress)

		if err == nil {
			log.Println("Successfully connected to Kafka.")
			break
		}

		log.Printf("Kafka not ready. Error: %v. Retrying in %s...\n", err, retryInterval)
		time.Sleep(retryInterval)
	}

	if conn == nil {
		return fmt.Errorf("failed to connect to Kafka after %d attempts (last error: %w)", retriesNumber, err)
	}
	defer conn.Close()

	// Connection established, so we can create the topic
	topicConfig := kafka.TopicConfig{
		Topic:             topicName,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	log.Printf("Attempting to create topic '%s'...", topicName)
	for i := range retriesNumber {
		err = conn.CreateTopics(topicConfig)

		if err == nil {
			log.Printf("Topic '%s' created successfully.", topicName)
			return nil
		}

		if errors.Is(err, kafka.TopicAlreadyExists) {
			log.Printf("Topic '%s' already exists. Skipping creation.", topicName)
			return nil
		}

		// Check for retryable race condition
		if errors.Is(err, kafka.InvalidReplicationFactor) {
			log.Printf("Kafka controller not ready, got 'InvalidReplicationFactor'. Retrying in %s seconds... (Attempt %d/%d)", retryInterval, i+1, retriesNumber)
			time.Sleep(retryInterval)
			continue
		}

		return fmt.Errorf("failed to create topic (non-retryable error): %w", err)
	}

	return fmt.Errorf("failed to create topic after %d attempts (last error: %w)", retriesNumber, err)
}

func generateEvent() Event {
	userID := fmt.Sprintf("user_%d", rand.IntN(10)+1)
	eventTypes := []string{"view", "click", "purchase", "add_to_cart", "login"}
	eventType := eventTypes[rand.IntN(len(eventTypes))]
	urls := []string{"/home", "/product/a", "/product/b", "/checkout", "/blog", "/about"}
	url := urls[rand.IntN(len(urls))]

	return Event{
		UserID:    userID,
		EventType: eventType,
		Timestamp: time.Now().UnixMilli(),
		URL:       url,
	}
}
