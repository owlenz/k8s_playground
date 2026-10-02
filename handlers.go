package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Name      string        `bson:"name"`
	Role      int           `bson:"role"`
	CreatedAt time.Time     `bson:"createdAt"`
}

type App struct {
	DB         *mongo.Client
	UsersCache usersCache
}

func (a *App) GetUserHandler(w http.ResponseWriter, r *http.Request) {

	data, ok := a.UsersCache.get()

	if !ok {
		v, err, _ := a.UsersCache.group.Do("users", func() (any, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			users := a.DB.Database("devdb").Collection("users")

			opts := options.Find().
				SetSort(bson.D{{Key: "_id", Value: 1}}).
				SetLimit(50)
			cur, err := users.Find(ctx, bson.D{}, opts)
			if err != nil {
				log.Printf("getting users: %v", err)
				return nil, err
			}
			defer cur.Close(ctx)

			var res []bson.M
			if err := cur.All(ctx, &res); err != nil {
				log.Printf("decoding users: %v", err)
				return nil, err
			}
			if res == nil {
				res = []bson.M{}
			}
			b, err := json.Marshal(res)
			if err != nil {
				return nil, err
			}
			a.UsersCache.set(b, 2*time.Second)
			return b, nil
		})
		if err != nil {
			log.Printf("getting users: %v", err)
			http.Error(w, "failed to get users", http.StatusInternalServerError)
			return
		}
		data = v.([]byte)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (a *App) SetUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var in User
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	users := a.DB.Database("devdb").Collection("users")
	res, err := users.InsertOne(ctx, in)
	if err != nil {
		log.Printf("insert user: %v", err)
		http.Error(w, "failed to save user", http.StatusInternalServerError)
		return
	}

	a.UsersCache.invalidate()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": res.InsertedID})
}
