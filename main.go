package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v9"
)

type RedisCli struct {
	r   *redis.Client
	ctx context.Context
}

type RedisEntry struct {
	keyName  string
	keyvalue string
}

const (
	interval  = 600
	redisAddr = "192.168.0.123:6379"
)

func main() {
	go storeContainerMetrics(interval)
	log.Println("started gathering container metrics...")

	go storeHostMetrics(interval)
	log.Println("started gathering host metrics...")

	stopChan := make(chan os.Signal, 2)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-stopChan
}

func newRedisCli(addr string, ctx context.Context) RedisCli {
	return RedisCli{
		r: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		}),
		ctx: ctx,
	}
}
