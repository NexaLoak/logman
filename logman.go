package main

import (
	"context"
	"github.com/redis/go-redis/v9"
)

const LogKey = "log"

type Logman struct {
	redisAddress  string
	redisPassword string
	redisDB       int
	redisClient   *redis.Client
}

func NewLogman(redisAddress, redisPassword string, redisDB int) *Logman {
	return &Logman{
		redisAddress:  redisAddress,
		redisPassword: redisPassword,
		redisDB:       redisDB,
		redisClient: redis.NewClient(&redis.Options{
			Addr:     redisAddress,
			Password: redisPassword,
			DB:       redisDB,
		}),
	}
}

func (l *Logman) checkRedisClient() {
	_, err := l.redisClient.Ping(context.Background()).Result()
	if err != nil {
		l.redisClient = redis.NewClient(&redis.Options{
			Addr:     l.redisAddress,
			Password: l.redisPassword,
			DB:       l.redisDB,
		})
	}
}

func (l *Logman) getContext(c ...context.Context) context.Context {
	if len(c) == 0 {
		return context.Background()
	}
	return c[0]
}

func (l *Logman) Push(log *Log, c ...context.Context) {
	l.checkRedisClient()
	ctx := l.getContext(c...)

	// TODO: Fetch id of last command
	l.redisClient.RPush(ctx, LogKey, log.Encode())
}

func main() {

}
