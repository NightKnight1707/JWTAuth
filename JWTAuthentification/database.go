package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RefreshToken struct {
	Guid    string    `bson:"_id"`
	Refresh string    `bson:"refresh"`
	Time    time.Time `bson:"time"`
}

var collection *mongo.Collection

func initDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		logrus.Error(err)
	}
	collection = client.Database("auth").Collection("tokens")
}

func ReadRefreshToken(guid string) (*RefreshToken, error) {
	filter := bson.D{{"_id", guid}}
	var result *RefreshToken
	if err := collection.FindOne(context.TODO(), filter).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func UpdateRefreshToken(refresh, guid string) error {
	filter := bson.D{{"_id", guid}}
	update := bson.D{
		{"$set", bson.D{
			{"refresh", refresh},
			{"time", time.Now().Add(60 * 24 * time.Hour)},
		}},
	}
	_, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func InsertRefreshToken(refresh, guid string) error {
	token := RefreshToken{Refresh: refresh, Guid: guid, Time: time.Now().Add(60 * 24 * time.Hour)}
	_, err := collection.InsertOne(context.TODO(), token)
	if err != nil {
		return err
	}
	return nil
}
