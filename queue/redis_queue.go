package queue

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	ctx    context.Context
	client *redis.Client
	key    string
}

type RedisConfig struct {
	// the redis url to connect to
	RedisURL string
	// the redis password to use
	RedisPassword string
	// the redis db to use
	RedisDB int
	// the redis key to use
	RedisKey string
}

// NewRedis returns a new scrapper instance
func NewRedis(ctx context.Context, config RedisConfig) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisURL,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})
	return &RedisQueue{
		ctx:    ctx,
		client: client,
	}, nil
}

// Close closes the underlying client object.
func (r *RedisQueue) Close() error {
	return r.client.Close()
}

// ReadMessage reads a single message from the underlying client object and then calls the callback function
// passing the message as a parameter.
// If the callback function returns an error, the message is put back in the queue.
func (r *RedisQueue) ReadMessage(callback func([]byte) error) ([]byte, error) {
	msg, err := r.client.BLPop(r.ctx, 0, r.key).Result()
	if err != nil {
		return nil, err
	}
	if callback == nil {
		return []byte(msg[1]), nil
	}
	err = callback([]byte(msg[1]))
	if err != nil {
		// put the message back in the queue
		err2 := r.client.LPush(r.ctx, r.key, msg).Err()
		if err2 != nil {
			return nil, fmt.Errorf("error putting message back in the queue: %v", err2)
		}
	}
	return []byte(msg[1]), nil
}

// WriteMessage writes a single message to the underlying client object and then calls the callback function.
func (r *RedisQueue) WriteMessage(msg []byte, callback func([]byte) error) error {
	err := r.client.RPush(r.ctx, r.key, msg).Err()
	if err != nil {
		return err
	}
	if callback == nil {
		return nil
	}
	return callback(msg)
}
