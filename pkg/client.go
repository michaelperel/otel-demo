package pkg

import (
	"github.com/go-redis/redis/v8"
)

type Client struct{ *redis.Client }

func NewClient(url string) (*Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	return &Client{redis.NewClient(opts)}, nil
}
