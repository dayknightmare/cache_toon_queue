package lifo

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/Vupy/cache-toon-queue/src/utils"
	"github.com/go-redis/redis"
)

func addLifoPriority(c *structs.RedisStruct, queueName string, message structs.ItemOptions) (*structs.QueueItem, error) {
	item := structs.NewQueueItem(message)
	out, err := json.Marshal(item)
	priority, _ := strconv.Atoi(strconv.Itoa(item.Priority) + strconv.Itoa(int(time.Now().UTC().Unix())))

	if err != nil {
		return nil, err
	}

	z := redis.Z{
		Score:  float64(priority),
		Member: out,
	}

	e := c.Client.ZAdd(utils.Sufixer(prefix+queueName, "lifo", 1), z)

	if e.Err() != nil {
		return nil, e.Err()
	}

	return item, nil
}

func getLifoPriority(c *structs.RedisStruct, queueName string) ([]redis.Z, error) {
	return c.Client.ZPopMin(utils.Sufixer(prefix+queueName, "lifo", 1)).Result()
}
