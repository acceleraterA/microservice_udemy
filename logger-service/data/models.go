package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Models struct {
	LogEntry LogEntry
}

type LogEntry struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	Data      string    `bson:"data" json:"data"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func New(mongo *mongo.Client) Models {
	client = mongo
	return Models{
		LogEntry: LogEntry{},
	}

}

func (l *LogEntry) Insert(entry LogEntry) error {
	// get the collection
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), LogEntry{
		Name:      entry.Name,
		Data:      entry.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Println("Error inserting log entry", err)
		return err
	}
	return nil
}

func (l *LogEntry) All() ([]*LogEntry, error) {
	// create a context
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// get the collection
	collection := client.Database("logs").Collection("logs")
	ops := options.Find()
	ops.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(context.TODO(), bson.D{}, ops)
	if err != nil {
		log.Println("Error getting log entries", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*LogEntry
	for cursor.Next(ctx) {
		var item LogEntry
		err := cursor.Decode(&item)
		if err != nil {
			log.Println("Error decoding entry", err)
			return nil, err
		}
		logs = append(logs, &item)
	}
	return logs, nil
}
