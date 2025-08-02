package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

var AppConfig *Config

type Config struct {
	RabbitMQURL           string
	QueueName             string
	RedisAddr             string
	RedisPassword         string
	RedisDB               int
	NumProducerRequests   int
	PublishInterval       int
	KeyPrefixReportStatus string
	KeyPrefixReportData   string
	WorkerTimeout         int
	NumWorkers            int
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("No .env file found or unable to load it, using environment variables...")
	}

	return &Config{
		RabbitMQURL:           getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		QueueName:             getEnv("QUEUE_NAME", "report_requests"),
		RedisAddr:             getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:         getEnv("REDIS_PASSWORD", ""),
		RedisDB:               getEnvInt("REDIS_DB"),
		NumProducerRequests:   getEnvInt("NUM_PRODUCER_REQUESTS"),
		PublishInterval:       getEnvInt("NUM_WORKERS"),
		KeyPrefixReportStatus: getEnv("KEY_PREFIX_REPORT_STATUS", "report:status:"),
		KeyPrefixReportData:   getEnv("KEY_PREFIX_REPORT_DATA", "report:data:"),
		WorkerTimeout:         getEnvInt("WORKER_TIMEOUT"),
		NumWorkers:            getEnvInt("PUBLISH_INTERVAL"),
	}
}

// getEnv mengambil environment variable, jika tidak ada pakai defaultValue
func getEnv(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}

func getEnvInt(key string) int {
	result, err := strconv.Atoi(getEnv(key, "0"))
	if err != nil {
		log.Fatalf("Invalid REDIS_DB: %v", err)
	}

	return result
}
