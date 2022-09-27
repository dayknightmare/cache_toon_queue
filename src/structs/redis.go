package structs

import (
	"github.com/go-redis/redis"
)

type RedisStruct struct {
	Client Commander
}

var RedisQueue *RedisStruct

func NewRedisStruct() *RedisStruct {
	r := &RedisStruct{}

	if OptionsLoaded.RedisIsCluster {
		r.Client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: OptionsLoaded.RedisClusterHosts,
		})
	} else {
		r.Client = redis.NewClient(&redis.Options{
			Addr: OptionsLoaded.RedisHost + ":" + OptionsLoaded.RedisPort,
		})
	}

	return r
}
