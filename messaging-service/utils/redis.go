package utils

import (
	"context"
	"encoding/json"
	"log"
	redisClient "messaging-service/redis"
	"messaging-service/types/requests"

	"github.com/redis/go-redis/v9"
)

func GetClientConnectionFromRedis(ctx context.Context, client *redisClient.RedisClient, userUUID string) (*requests.Connection, error) {
	var connection *requests.Connection
	err := client.Get(ctx, userUUID, connection)
	if err != nil {
		return nil, err
	}

	if connection == nil {
		return nil, nil
	}

	return connection, nil
}

func SetClientConnectionToRedis(ctx context.Context, client *redisClient.RedisClient, connection *requests.Connection) error {
	return client.Set(ctx, connection.UserUUID, connection)
}

// subscribe to the channel
func SubscribeToChannel(subscriber *redis.PubSub, fn func(event string) error) {
	for redisMsg := range subscriber.Channel() {
		err := fn(redisMsg.Payload)
		if err != nil {
			log.Println(err)
		}
	}
}

// pass in the identifier of the channel so Redis can perform pub/sub
// TODO – use handler context
func SetupChannel(c *redisClient.RedisClient, channelName string) *redis.PubSub {
	subscriber := c.Client.Subscribe(context.Background(), channelName)
	return subscriber
}

// TODO use context from handler
func PublishToRedisChannel(c *redisClient.RedisClient, channelName string, v interface{}) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}

	res := c.Client.Publish(context.Background(), channelName, bytes)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}
