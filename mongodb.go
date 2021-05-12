package main

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func mongoDB() *mongo.Database {
	dbOnce.Do(func() {
		client, err := newMongoDB()
		if err != nil {
			log.Fatal(err)
		}
		clientDB = client
	})
	return clientDB.Database("registry")
}

func newMongoDB() (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://" + *mongodbUsername + ":" + *mongoDBpassword + "@" + *mongoDBHostAddr))
	if err != nil {
		log.Fatalf("MongoDB NewClient failed: %s", err)
		return nil, err
	}

	if err := connectMongoDB(client); err != nil {
		return nil, err
	}

	if err := pingMongoDB(client); err != nil {
		return nil, err
	}

	return client, nil
}

func connectMongoDB(client *mongo.Client) error {
	ctx, cancel := operationTimeout(defaultMongoDBOpTimeout)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		log.Warnf("MongoDB connection failed: %s", err)
		return err
	}
	return nil
}

func pingMongoDB(client *mongo.Client) error {
	ctx, cancel := operationTimeout(defaultMongoDBOpTimeout)
	defer cancel()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Warnf("MongoDB ping failed: %s", err)
		return err
	}
	return nil
}

func operationTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
