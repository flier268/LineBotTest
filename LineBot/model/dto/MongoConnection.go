package dto

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnection struct {
	client *mongo.Client
}

func (connection *MongoConnection) Connect(url string) (IConnection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(url))

	if err == nil {
		return &MongoConnection{c}, nil
	}
	return nil, err
}
func (connection *MongoConnection) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer func() {
		if err := connection.client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func (connection *MongoConnection) Get(userID string) any {
	collection := connection.client.Database("linebot").Collection("messageReceived")
	filter := bson.D{{Key: "userId", Value: userID}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	type r struct {
		Context string
		Type    string
		Time    int
	}
	buffer := make([]r, 0)

	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result bson.D
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		// log.Println(result)
		data := new(r)
		for _, pair := range result {
			switch pair.Key {
			case "context":
				data.Context = pair.Value.(string)
			case "type":
				data.Type = pair.Value.(string)
			case "time":
				data.Time = int(pair.Value.(primitive.DateTime).Time().Unix())
			}
		}

		buffer = append(buffer, *data)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	return buffer
}

func (connection *MongoConnection) Insert(model MessageModel) (interface{}, error) {
	collection := connection.client.Database("linebot").Collection("messageReceived")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := collection.InsertOne(ctx, bson.D{{Key: "userId", Value: model.UserID}, {Key: "context", Value: model.Context}, {Key: "type", Value: model.Type}, {Key: "time", Value: model.Time}})
	id := res.InsertedID
	return id, err
}
