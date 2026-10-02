package main

import (
	"context"
	"fmt"

	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	// mongodb connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mongoClient := mongodbConnect()

	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting: %v", err)
		}
	}()
	app := &App{
		DB: mongoClient,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", app.GetUserHandler)
	mux.HandleFunc("POST /users", app.SetUserHandler)

	http.ListenAndServe(":4567", mux)
}

func mongodbConnect() *mongo.Client {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://root:secretpassword@localhost:27017"
	}

	clientOpts := options.Client().ApplyURI(mongoURI)

	var client *mongo.Client
	var err error

	for attempt := 1; attempt <= 10; attempt++ {
		client, err = mongo.Connect(clientOpts)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			err = client.Ping(ctx, nil)
			cancel()
			if err == nil {
				break
			}
			_ = client.Disconnect(context.Background())
		}
		log.Printf("mongo not ready (attempt %d): %v", attempt, err)
		time.Sleep(time.Duration(attempt) * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	fmt.Println("Successfully connected to MongoDB!")
	return client
}
