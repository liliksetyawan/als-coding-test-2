package services

import (
	"als-coding-test-2/config"
	"als-coding-test-2/constanta"
	"als-coding-test-2/models"
	"als-coding-test-2/server"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"log"
	"math/rand"
	"sync"
	"time"
)

func ProduceReportRequests(ctx context.Context) {
	ch := server.GetServerAttribute().RabbitMQChan

	for i := 0; i < config.AppConfig.NumProducerRequests; i++ {
		select {
		case <-ctx.Done():
			log.Println("[Producer] Context cancelled, stopping producer")
			return
		default:
			req := models.ReportRequest{
				ID:         fmt.Sprintf("report-%d", i),
				ReportType: "sales",
				Parameters: map[string]string{"region": "ID"},
				CreatedAt:  time.Now(),
			}
			body, _ := json.Marshal(req)
			err := ch.Publish("", config.AppConfig.QueueName, false, false, amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         body,
			})
			if err != nil {
				log.Printf("[Producer] Failed to publish message: %v", err)
			}
			time.Sleep(time.Duration(config.AppConfig.PublishInterval) * time.Millisecond)
		}
	}
}

func UpdateReportStatus(ctx context.Context, rdb *redis.Client, requestID string, status constanta.ReportStatus, reportData string, errMsg string) {
	statusKey := config.AppConfig.KeyPrefixReportStatus + requestID
	rdb.Set(ctx, statusKey, string(status), 0)
	if status == constanta.StatusCompleted || status == constanta.StatusFailed {
		result := models.ReportResult{
			RequestID:   requestID,
			Status:      status,
			GeneratedAt: time.Now(),
			ReportData:  reportData,
			Error:       errMsg,
		}
		dataKey := config.AppConfig.KeyPrefixReportData + requestID
		bytes, _ := json.Marshal(result)
		rdb.Set(ctx, dataKey, bytes, 0)
	}
}

func ReportWorker(ctx context.Context, workerID int, tasks <-chan amqp.Delivery, results chan<- models.ReportResult, rdb *redis.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-tasks:
			if !ok {
				return
			}
			var req models.ReportRequest
			if err := json.Unmarshal(d.Body, &req); err != nil {
				log.Printf("[Worker %d] Failed to unmarshal: %v", workerID, err)
				continue
			}
			UpdateReportStatus(ctx, rdb, req.ID, constanta.StatusInProgress, "", "")

			childCtx, cancel := context.WithTimeout(ctx, time.Duration(config.AppConfig.WorkerTimeout)*time.Second)
			data, err := simulateReportGeneration(childCtx, req)
			cancel()

			res := models.ReportResult{
				RequestID:   req.ID,
				GeneratedAt: time.Now(),
				Status:      constanta.StatusCompleted,
				ReportData:  data,
			}
			if err != nil {
				res.Status = constanta.StatusFailed
				res.Error = err.Error()
			}
			UpdateReportStatus(ctx, rdb, req.ID, res.Status, res.ReportData, res.Error)
			results <- res
		}
	}
}

func ResultAckHandler(ctx context.Context, results <-chan models.ReportResult, deliveries map[string]amqp.Delivery, mu *sync.Mutex) {
	for {
		select {
		case <-ctx.Done():
			return
		case res := <-results:
			mu.Lock()
			d, ok := deliveries[res.RequestID]
			if ok {
				if res.Status == constanta.StatusCompleted || res.Status == constanta.StatusFailed {
					d.Ack(false)
				}
				delete(deliveries, res.RequestID)
			}
			mu.Unlock()
		}
	}
}

func StartReportProcessor(ctx context.Context, rdb *redis.Client) {
	ch := server.GetServerAttribute().RabbitMQChan
	numberWorker := config.AppConfig.NumWorkers
	queueName := config.AppConfig.QueueName

	deliveriesMap := make(map[string]amqp.Delivery)
	mu := sync.Mutex{}
	tasks := make(chan amqp.Delivery, numberWorker)
	results := make(chan models.ReportResult, numberWorker)

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	for i := 0; i < numberWorker; i++ {
		go ReportWorker(ctx, i, tasks, results, rdb)
	}
	go ResultAckHandler(ctx, results, deliveriesMap, &mu)

	for {
		select {
		case <-ctx.Done():
			close(tasks)
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			var req models.ReportRequest
			if err := json.Unmarshal(d.Body, &req); err == nil {
				UpdateReportStatus(ctx, rdb, req.ID, constanta.StatusPending, "", "")
				mu.Lock()
				deliveriesMap[req.ID] = d
				mu.Unlock()
				tasks <- d
			}
		}
	}

}

// simulateReportGeneration simulates the actual report generation process.
// It includes random delays and a chance of failure.
// It also respects its own context for timeout/cancellation.
func simulateReportGeneration(ctx context.Context, request models.ReportRequest) (string, error) {
	log.Printf("[Worker: %s] Starting report generation for ID: %s (Type: %s)", request.ID, request.ID,
		request.ReportType)
	// Simulate CPU-intensive work or external API calls
	delay := time.Duration(1+rand.Intn(5)) * time.Second // 1 to 5 seconds
	select {
	case <-time.After(delay):
		// Continue processing
	case <-ctx.Done():
		log.Printf("[Worker: %s] Context cancelled during simulation for ID: %s", request.ID, request.ID)
		return "", ctx.Err() // Propagate context cancellation error
	}
	// Simulate random failure (e.g., database error, invalid parameters)
	if rand.Intn(100) < 20 { // 20% chance of failure
		log.Printf("[Worker: %s] Simulated failure for ID: %s", request.ID, request.ID)
		return "", fmt.Errorf("simulated report generation error for ID %s", request.ID)
	}
	reportData := fmt.Sprintf("Report %s - Type: %s, Generated On: %s, Data: Random Value %d",
		request.ID, request.ReportType, time.Now().Format(time.RFC3339), rand.Intn(1000))
	log.Printf("[Worker: %s] Successfully generated report for ID: %s", request.ID, request.ID)
	return reportData, nil
}
