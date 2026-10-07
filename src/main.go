package main

import (
	"context"
	"log/slog"

	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	// mongodb connection
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mongoClient := mongodbConnect()

	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			slog.Error("Error disconnecting: ", "err", err)
			os.Exit(1)
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
		slog.Error("MONGO_URI isn't set")
		os.Exit(1)
	}

	slog.Warn("MONGO_URI ", "uri", mongoURI)

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
		slog.Warn("mongo not ready", "attempt", attempt, "err", err)
		time.Sleep(time.Duration(attempt) * time.Second)
	}
	if err != nil {
		slog.Error("Failed to connect to MongoDB", "err", err)
		os.Exit(1)
	}
	slog.Info("Successfully connected to MongoDB!")
	return client
}
