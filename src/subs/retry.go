package subs

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/Vupy/cache-toon-queue/src/app"
	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/go-redis/redis"
)

func neverRetry(msg string, key string) {
	z := redis.Z{
		Score:  float64(-1),
		Member: msg,
	}

	structs.RedisQueue.Client.ZAdd(
		key,
		z,
	)
}

func removeMessage(msg string, key string) {
	structs.RedisQueue.Client.ZRem(key, msg)
}

func retryMessages(msgs []string, key string) {
	for _, msg := range msgs {
		var queueItem structs.QueueItem

		removeMessage(msg, key)

		err := json.Unmarshal([]byte(msg), &queueItem)

		if err != nil {
			continue
		}

		if queueItem.Attempt == 0 {
			neverRetry(msg, key)
			continue
		}

		queue := strings.Split(key, ":")[1]

		newMessage := structs.ItemOptions{
			Value:     queueItem.Value,
			Attempt:   queueItem.Attempt - 1,
			TypeQueue: queueItem.TypeQueue,
			Priority:  queueItem.Priority,
		}

		app.AddMessage(
			queue,
			newMessage,
		)
	}
}

func getMessages(keys []string) {
	max := time.Now()
	min := max.Add(time.Minute * -10)

	score := redis.ZRangeBy{
		Min: strconv.Itoa(int(min.UnixMilli())),
		Max: strconv.Itoa(int(max.UnixMilli())),
	}

	for _, key := range keys {
		msgs, err := structs.RedisQueue.Client.ZRangeByScore(key, score).Result()

		if err != nil {
			panic(err)
		}

		retryMessages(msgs, key)
	}
}

func searchErrorMessages() {
	keys, err := structs.RedisQueue.Client.Keys("error:*").Result()

	if err != nil {
		panic(err)
	}

	getMessages(keys)
}

func StartRetrier() {
	for {
		searchErrorMessages()
		time.Sleep(time.Second * 30)
	}
}
