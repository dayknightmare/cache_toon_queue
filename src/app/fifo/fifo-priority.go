package fifo

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/Vupy/cache-toon-queue/src/utils"
	"github.com/go-redis/redis"
)

func addFifoPriority(c *structs.RedisStruct, queueName string, message structs.ItemOptions) (*structs.QueueItem, error) {
	item := structs.NewQueueItem(message)
	out, err := json.Marshal(item)
	priority, _ := strconv.Atoi(strconv.Itoa(item.Priority) + strconv.Itoa(int(time.Now().UTC().UnixMilli())))

	if err != nil {
		return nil, err
	}

	z := redis.Z{
		Score:  float64(priority),
		Member: out,
	}

	e := c.Client.ZAdd(utils.Sufixer(prefix+queueName, "fifo", 1), z)

	if e.Err() != nil {
		return nil, e.Err()
	}

	return item, nil
}

func getFifoPriority(c *structs.RedisStruct, queueName string) ([]redis.Z, error) {
	return c.Client.ZPopMin(utils.Sufixer(prefix+queueName, "fifo", 1)).Result()
}
