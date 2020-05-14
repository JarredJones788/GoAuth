package cache

import (
	"errors"
	"fmt"
	"types"

	"github.com/go-redis/redis"
)

//Cache - cache
type Cache struct {
	Client  *redis.Client
	Timeout int
	Active  bool
}

//Init - connect to cache server
func (cache Cache) Init(config *types.Config) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr: config.Redis.Conn,
		DB:   0,
	})
	cache.Client = client
	cache.Timeout = config.Redis.Timeout
	cache.Active = config.Redis.Active

	if cache.Active {
		_, err := client.Ping().Result()
		if err != nil {
			fmt.Println(err)
			return &cache
		}
	}

	return &cache
}

//Set - sets a key -> value pair
func (cache Cache) Set(key string, value string) error {
	if !cache.Active {
		return nil
	}
	err := cache.Client.Set(key, value, 0).Err()
	cache.Client.Do("EXPIRE", key, cache.Timeout)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

//Get - gets a value from a key.
func (cache Cache) Get(key string) (string, error) {
	if !cache.Active {
		return "", errors.New("Cache disabled")
	}
	val, err := cache.Client.Get(key).Result()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return val, nil
}

//Del - removes item from cache
func (cache Cache) Del(key string) error {
	if !cache.Active {
		return errors.New("Cache disabled")
	}
	cache.Client.Del(key)
	return nil
}
