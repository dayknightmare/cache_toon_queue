package lifo

import (
	"encoding/json"

	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/Vupy/cache-toon-queue/src/utils"
)

const prefix = "cache_toon_"

func addLifo(c *structs.RedisStruct, queueName string, message structs.ItemOptions) (*structs.QueueItem, error) {
	item := structs.NewQueueItem(message)
	out, err := json.Marshal(item)

	if err != nil {
		return nil, err
	}

	e := c.Client.LPush(utils.Sufixer(prefix+queueName, "lifo", 0), out)
	if e.Err() != nil {
		return nil, e.Err()
	}

	return item, nil
}

func getLifo(c *structs.RedisStruct, queueName string) (string, error) {
	return c.Client.LPop(utils.Sufixer(prefix+queueName, "lifo", 0)).Result()
}
