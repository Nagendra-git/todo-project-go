// Command server starts the Todo HTTP API.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/todo-go-app/internal/config"
	"github.com/example/todo-go-app/internal/db"
	"github.com/example/todo-go-app/internal/handlers"
	"github.com/example/todo-go-app/internal/repository"
	"github.com/example/todo-go-app/internal/router"
)

func main() {
	cfg, err := config.Load("configs/application.properties")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	log.Println("CORS Origin:", cfg.CORS.AllowedOrigin)
	client, err := db.Connect(cfg.Mongo.URI, cfg.Mongo.Timeout)
	if err != nil {
		log.Fatalf("failed to connect to mongo: %v", err)
	}
	defer func() {
		if err := db.Disconnect(client, cfg.Mongo.Timeout); err != nil {
			log.Printf("error disconnecting from mongo: %v", err)
		}
	}()
	log.Println("connected to MongoDB at", cfg.Mongo.URI)

	collection := client.Database(cfg.Mongo.Database).Collection(cfg.Mongo.Collection)
	todoRepo := repository.NewMongoTodoRepository(collection)
	todoHandler := handlers.NewTodoHandler(todoRepo)

	handler := router.New(todoHandler, cfg.CORS.AllowedOrigin)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Run the server in a goroutine so we can listen for shutdown signals.
	go func() {
		log.Println("server listening on port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM (e.g. Ctrl+C, container stop).
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("error during shutdown: %v", err)
	}
	log.Println("server stopped")
}
