package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"qq-bot/config"
	"qq-bot/utils"
	"time"
)

var (
	mongoClient *mongo.Client
	dbName      string
)

func InitMongo() {
	var uri string
	if config.Config.Database.Mongo.User != "" && config.Config.Database.Mongo.Pwd != "" {
		uri = fmt.Sprintf(
			"mongodb://%s:%s@%s:%d/%s?w=majority&authSource=admin",
			config.Config.Database.Mongo.User,
			config.Config.Database.Mongo.Pwd,
			config.Config.Database.Mongo.Hostname,
			config.Config.Database.Mongo.Port,
			config.Config.Database.Mongo.Database,
		)
	} else {
		uri = fmt.Sprintf(
			"mongodb://%s:%d/%s?w=majority",
			config.Config.Database.Mongo.Hostname,
			config.Config.Database.Mongo.Port,
			config.Config.Database.Mongo.Database,
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	utils.PanicNotNil(err)
	mongoClient = client
	dbName = config.Config.Database.Mongo.Database
}

func Test() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	utils.PanicNotNil(mongoClient.Ping(ctx, nil))
}

func Collection(collectionName string) *mongo.Collection {
	return mongoClient.Database(dbName).Collection(collectionName)
}
