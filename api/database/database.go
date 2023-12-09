package database

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()

func CreateClient(dbNo int) *redis.Client {
	var redisAddress string = os.Getenv("DB_ADDR")
	var redisPassword string = os.Getenv("DB_PASS")
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword,
		DB:       dbNo,
	})

	return rdb
}
