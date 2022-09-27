package lifo

import (
	"encoding/json"

	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/Vupy/cache-toon-queue/src/utils"
)

type Lifo struct{}

func (l *Lifo) AddMessage(queueName string, message structs.ItemOptions) (*structs.QueueItem, error) {
	message = utils.FixMessage(message)

	if message.Priority == 0 {
		return addLifo(structs.RedisQueue, queueName, message)
	}

	return addLifoPriority(structs.RedisQueue, queueName, message)
}

func (l *Lifo) GetMessage(queueName string) (structs.QueueItem, error) {
	item, err := getLifoPriority(structs.RedisQueue, queueName)
	var message structs.QueueItem

	if err != nil || len(item) == 0 {
		itemFifo, err := getLifo(structs.RedisQueue, queueName)

		if err != nil || len(itemFifo) == 0 {
			return structs.QueueItem{}, err
		}

		text := itemFifo
		err = json.Unmarshal([]byte(text), &message)

		return message, err
	}

	text := item[0].Member.(string)
	err = json.Unmarshal([]byte(text), &message)

	return message, err
}
