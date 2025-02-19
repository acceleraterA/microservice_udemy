package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb://localhost:27017"
	gRPCPort = "50051"
)

var client *mongo.Client

type Config struct {
}

func main() {

	// connect to mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic("Failed to connect to mongo")
	}
	client = mongoClient
	// create a context in order to disconnect from mongo
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) //create a context that will cancel after 15 seconds
	defer cancel()                                                           //cancel the context when the function returns

	defer func() { //defer a function to disconnect from mongo at the end of main function
		if err := client.Disconnect(ctx); err != nil {
			log.Panic("Failed to disconnect from mongo")
		}
	}()
}

func connectToMongo() (*mongo.Client, error) {
	//create connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})
	//connect to mongo
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error connecting to MongoDB", err)
		return nil, err
	}
	log.Println("Connected to MongoDB")
	return c, nil
}
