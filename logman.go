package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"sync"
)

const LogKey = "log"

var wg = sync.WaitGroup{}

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

	if len(result) == 0 {
		return -1
	}

	lastEncodedLog := result[0]
	lastLog := DecodeLog(lastEncodedLog)
	return lastLog.id
}

func (l *Logman) fetchLastLogCommitIndex(c ...context.Context) int {
	l.checkRedisClient()
	ctx := l.getContext(c...)

	result, err := l.redisClient.LRange(ctx, LogKey, -1, -1).Result()
	if err != nil {
		log.Fatal("An error occurred in fetchLastLogCommitIndex() method while getting last log: ", err)
	}

	if len(result) == 0 {
		return -1
	}

	lastEncodedLog := result[0]
	lastLog := DecodeLog(lastEncodedLog)
	return lastLog.commitIndex
}

func (l *Logman) Push(log *Log, c ...context.Context) {
	l.checkRedisClient()
	ctx := l.getContext(c...)

	wg.Wait()

	log.id = l.fetchLastLogId(c...) + 1
	log.commitIndex = l.fetchLastLogCommitIndex(c...)
	l.redisClient.RPush(ctx, LogKey, log.Encode())
}

func (l *Logman) hardPush(log *Log, c ...context.Context) {
	l.checkRedisClient()
	ctx := l.getContext(c...)

	log.id = l.fetchLastLogId(c...) + 1
	l.redisClient.RPush(ctx, LogKey, log.Encode())
}

func (l *Logman) RetrieveLastLog(c ...context.Context) *Log {
	l.checkRedisClient()
	ctx := l.getContext(c...)

	result, err := l.redisClient.LRange(ctx, LogKey, -1, -1).Result()
	if err != nil {
		log.Fatal("An error occurred in RetrieveLastLog() method while getting last log: ", err)
	}

	lastEncodedLog := result[0]
	lastLog := DecodeLog(lastEncodedLog)
	return lastLog
}

func (l *Logman) RetrieveLogs(start, end int64, c ...context.Context) []Log {
	// TODO: Return all if something is persisted on dist
	l.checkRedisClient()
	ctx := l.getContext(c...)

	result, err := l.redisClient.LRange(ctx, LogKey, start, end).Result()
	if err != nil {
		log.Fatal("An error occurred in RetrieveLogs() method while getting logs: ", err)
	}

	logs := make([]Log, len(result))
	for i, encodedLog := range result {
		decodedLog := DecodeLog(encodedLog)
		logs[i] = *decodedLog
	}

	return logs
}

func (l *Logman) RetrieveAllLogs(c ...context.Context) []Log {
	return l.RetrieveLogs(0, -1, c...)
}

func (l *Logman) RetrieveUncommittedLogs(c ...context.Context) []Log {
	return l.RetrieveLogs(int64(l.fetchLastLogCommitIndex(c...))+1, int64(-1), c...)
}

func (l *Logman) trim(start, end int64, c ...context.Context) {
	l.checkRedisClient()
	ctx := l.getContext(c...)

	_, err := l.redisClient.LTrim(ctx, LogKey, start, end).Result()
	if err != nil {
		log.Fatal("An error occurred in Trim() method while trimming: ", err)
	}
}

func (l *Logman) TrimCommittedLogs(c ...context.Context) {
	lastCommittedIndex := int64(l.fetchLastLogCommitIndex(c...))
	if lastCommittedIndex == -1 {
		l.flushMemory(c...)
		return
	}
	l.trim(0, lastCommittedIndex)
}

func (l *Logman) flushMemory(c ...context.Context) {
	l.checkRedisClient()
	ctx := l.getContext(c...)

	l.redisClient.Del(ctx, LogKey)
}

func (l *Logman) Commit(c ...context.Context) {
	l.checkRedisClient()
	ctx := l.getContext(c...)

	defer wg.Done()
	wg.Add(1)

	uncommittedLogs := l.RetrieveUncommittedLogs(ctx)
	l.TrimCommittedLogs(ctx)

	newCommitIndex := uncommittedLogs[len(uncommittedLogs)-1].id
	for _, lg := range uncommittedLogs {
		lg.commitIndex = newCommitIndex
		lg.command.Execute()
		l.hardPush(&lg, ctx)
	}
}

func main() {
	lgman := NewLogman("127.0.0.1:6379", "", 0)
	ctx := context.Background()

	lgman.flushMemory(ctx)
	log1 := NewLog(10, 1, NewCommand(ShellClass, "ls"), 0)
	lgman.Push(log1)
	//fmt.Println(lgman.RetrieveLastLog().ToString())

	log2 := NewLog(10, 1, NewCommand(ShellClass, "uname -a"), 0)
	lgman.Push(log2)
	//fmt.Println(lgman.RetrieveLastLog().ToString())

	log3 := NewLog(10, 2, NewCommand(ShellClass, "dig A +short google.com"), 0)
	lgman.Push(log3)
	logs := lgman.RetrieveAllLogs(ctx)
	for _, lg := range logs {
		fmt.Println(lg.ToString())
	}

	lgman.Commit()
	logs = lgman.RetrieveAllLogs(ctx)
	for _, lg := range logs {
		fmt.Println(lg.ToString())
	}

	// TODO: Make a timeout code
	log4 := NewLog(10, 2, NewCommand(ShellClass, "echo Hey"), 0)
	lgman.Push(log4)

	log5 := NewLog(10, 2, NewCommand(ShellClass, "whoami"), 0)
	lgman.Push(log5)
	logs = lgman.RetrieveAllLogs(ctx)
	for _, lg := range logs {
		fmt.Println(lg.ToString())
	}

	lgman.Commit()
	logs = lgman.RetrieveAllLogs(ctx)
	for _, lg := range logs {
		fmt.Println(lg.ToString())
	}
}
