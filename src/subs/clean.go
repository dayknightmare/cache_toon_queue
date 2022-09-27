package subs

import (
	"strconv"
	"time"

	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/go-redis/redis"
)

func cleanCompletedMessages(msgs []string, key string) {
	for _, msg := range msgs {
		removeMessage(msg, key)
	}
}

func getCompletedMessages(keys []string) {
	max := time.Now().Add(time.Minute * -10)

	score := redis.ZRangeBy{
		Min: "0",
		Max: strconv.Itoa(int(max.UnixMilli())),
	}

	for _, key := range keys {
		msgs, err := structs.RedisQueue.Client.ZRangeByScore(key, score).Result()

		if err != nil {
			panic(err)
		}

		cleanCompletedMessages(msgs, key)
	}
}

func searchCompletedMessages() {
	keys, err := structs.RedisQueue.Client.Keys("completed:*").Result()

	if err != nil {
		panic(err)
	}

	getCompletedMessages(keys)
}

func StartCleaner() {
	for {
		searchCompletedMessages()
		time.Sleep(time.Second * 30)
	}
}
