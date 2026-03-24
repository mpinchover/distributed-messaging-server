package repo

import (
	"context"
	"fmt"
	"messaging-service/src/types/records"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	PAGINATION_MESSAGES = 20
	PAGINATION_ROOMS    = 10
)

type RepoInterface interface {
	// messaging
	LeaveRoom(userUUID string, roomUUID string) error
	GetMembersByRoomUUID(roomUUID string) ([]*records.Member, error)
	GetRoomByRoomUUID(roomUUID string) (*records.Room, error)
	SaveMessage(msg *records.Message) error
	GetMessagesByRoomUUID(roomUUID string, offset int) ([]*records.Message, error)
	GetMessagesByRoomUUIDs(roomUUIDs string, offset int) ([]*records.Message, error)
	GetRoomsByUserUUID(uuid string, offset int) ([]*records.Room, error)
	DeleteRoom(roomUUID string) error
	SaveRoom(room *records.Room) error
	GetRoomsByUserUUIDForSubscribing(userUUID string) ([]*records.Room, error)
}

type Repo struct {
	client   *mongo.Client
	db       *mongo.Database
	rooms    *mongo.Collection
	members  *mongo.Collection
	messages *mongo.Collection
}

func connect() (*mongo.Client, error) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		host := os.Getenv("MONGO_HOST")
		port := os.Getenv("MONGO_PORT")
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "27017"
		}
		uri = fmt.Sprintf("mongodb://%s:%s", host, port)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return mongo.Connect(ctx, options.Client().ApplyURI(uri))
}

func New() *Repo {
	client, err := connect()
	if err != nil {
		panic(err)
	}
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "messaging"
	}
	db := client.Database(dbName)
	r := &Repo{
		client:   client,
		db:       db,
		rooms:    db.Collection("rooms"),
		members:  db.Collection("members"),
		messages: db.Collection("messages"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := r.ensureIndexes(ctx); err != nil {
		panic(err)
	}
	return r
}

func (r *Repo) ensureIndexes(ctx context.Context) error {
	_, err := r.rooms.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "uuid", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	_, err = r.messages.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "uuid", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	_, err = r.members.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "room_uuid", Value: 1}, {Key: "user_uuid", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}
