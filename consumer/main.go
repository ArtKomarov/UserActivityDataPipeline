package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Event struct {
	UserID    string `json:"user_id" bson:"user_id"`
	EventType string `json:"event_type" bson:"event_type"`
	Timestamp int64  `json:"timestamp" bson:"timestamp"`
	URL       string `json:"url" bson:"url"`
}

const (
	topic          = "user_events"
	kafkaBroker    = "kafka:19092"
	mongoURI       = "mongodb://user:password@mongodb:27017"
	databaseName   = "user_data_db"
	collectionName = "events"
)

var (
	eventsConsumed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_consumer_events_total",
			Help: "Total number of user events consumed from Kafka.",
		},
	)
)

func setupMetrics() {
	prometheus.MustRegister(eventsConsumed)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	ctx := context.Background()

	mongoClient, err := setupMongoDB(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()
	collection := mongoClient.Database(databaseName).Collection(collectionName)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{kafkaBroker},
		Topic:       topic,
		GroupID:     "consumer-group-1",
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	log.Println("Starting metrics server on :8080")
	go setupMetrics()

	log.Println("Kafka Consumer started. Listening for events...")

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		var event Event
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Failed to unmarshal JSON: %v, raw message: %s", err, string(m.Value))
			continue
		}

		ingestionTime := time.Now().UnixMilli()
		processingTime := time.Now().UnixMilli()
		latencyMs := processingTime - event.Timestamp

		// Store the event in MongoDB
		result, err := collection.InsertOne(ctx, bson.M{
			"user_id":         event.UserID,
			"event_type":      event.EventType,
			"timestamp":       event.Timestamp,
			"url":             event.URL,
			"processing_time": ingestionTime,
			"latency_ms":      latencyMs,
		})

		if err != nil {
			log.Printf("Failed to insert document into MongoDB: %v", err)
		} else {
			eventsConsumed.Inc()
			log.Printf("Consumed event: UserID=%s, Type=%s. Stored in Mongo ID: %v", event.UserID, event.EventType, result.InsertedID)
		}
	}
}

// setupMongoDB connects to MongoDB and returns the client.
func setupMongoDB(ctx context.Context) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoURI)
	clientOptions.SetConnectTimeout(10 * time.Second)
	clientOptions.SetServerSelectionTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to MongoDB!")
	return client, nil
}
