package server

import (
	"als-coding-test-2/config"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"log"
	"sync"
)

type ServerAttribute struct {
	RabbitMQConn *amqp.Connection
	RabbitMQChan *amqp.Channel
	RedisClient  *redis.Client
}

var (
	instance *ServerAttribute
	once     sync.Once
)

func Initialize() {

	once.Do(func() {
		ctx := context.Background()

		redisClient, err := connectRedis(ctx)
		if err != nil {
			log.Fatalf("Failed to connect to Redis: %v", err)
		}

		rabbitConn, rabbitCh, err := connectRabbitMQ(ctx)
		if err != nil {
			log.Fatalf("Failed to initialize RabbitMQ: %v", err)
		}

		instance = &ServerAttribute{
			RabbitMQConn: rabbitConn,
			RabbitMQChan: rabbitCh,
			RedisClient:  redisClient,
		}
	})
}

func GetServerAttribute() *ServerAttribute {
	if instance == nil {
		log.Fatal("ServerAttribute not initialized. Call Initialize() first.")
	}
	return instance
}

// --- Fungsi Pembantu (Untuk Simulasi dan Koneksi) ---
// connectRabbitMQ establishes a connection and channel to RabbitMQ.
func connectRabbitMQ(ctx context.Context) (*amqp.Connection, *amqp.Channel, error) {
	queueName := config.AppConfig.QueueName

	conn, err := amqp.Dial(config.AppConfig.RabbitMQURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to open a channel: %w", err)
	}
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused

		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, fmt.Errorf("failed to declare a queue: %w", err)
	}
	log.Println("Successfully connected to RabbitMQ and declared queue:", queueName)
	return conn, ch, nil
}

// connectRedis establishes a connection to Redis.
func connectRedis(ctx context.Context) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: config.AppConfig.RedisAddr,
		DB:   0, // use default DB
	})
	// Use the provided context for the Ping operation
	err := rdb.Ping(ctx).Err()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
		return nil, err
	}
	log.Println("Successfully connected to Redis:", config.AppConfig.RedisAddr)
	return rdb, nil
}

func (s *ServerAttribute) Close() {
	if s.RabbitMQChan != nil {
		s.RabbitMQChan.Close()
	}
	if s.RabbitMQConn != nil {
		s.RabbitMQConn.Close()
	}
	if s.RedisClient != nil {
		s.RedisClient.Close()
	}
}
