package main

import (
	"als-coding-test-2/config"
	"als-coding-test-2/server"
	"als-coding-test-2/services"
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {

	config.AppConfig = config.LoadConfig()

	// Set ServerAttribute
	server.Initialize()

	defer func() {
		if err := recover(); err != nil {
			log.Println("Recovered from panic:", err)
		}
		if sa := server.GetServerAttribute(); sa != nil {
			sa.Close()
		}
	}()

	rand.Seed(time.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("Received shutdown signal")
		cancel()
	}()

	wg.Add(2)
	go func() {
		defer wg.Done()
		services.ProduceReportRequests(ctx)
	}()

	go func() {
		defer wg.Done()
		services.StartReportProcessor(ctx, server.GetServerAttribute().RedisClient)
	}()

	wg.Wait()

	log.Println("All processes stopped gracefully.")
}
