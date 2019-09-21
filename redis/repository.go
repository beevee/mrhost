package redis

import (
	"mrhost"

	"github.com/go-redis/redis"
)

type Repository struct {
	client *redis.Client
}

func NewRepository(addr string) mrhost.Repository {
	return Repository{client: redis.NewClient(&redis.Options{Addr: addr})}
}
