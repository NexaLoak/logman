package main

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
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

func (l *Logman) fetchLastLogId(c ...context.Context) int {
	l.checkRedisClient()
	ctx := l.getContext(c...)

	result, err := l.redisClient.LRange(ctx, LogKey, -1, -1).Result()
	if err != nil {
		log.Fatal("An error occurred in fetchLastLogId() method while getting last log: ", err)
	}

	lastEncodedLog := result[0]
	lastLog := DecodeLog(lastEncodedLog)
	return lastLog.id
}

func (l *Logman) Push(log *Log, c ...context.Context) {
	l.checkRedisClient()
	ctx := l.getContext(c...)

	log.id = l.fetchLastLogId(c...)
	l.redisClient.RPush(ctx, LogKey, log.Encode())
}

func main() {

}
