package fifo

import (
	"encoding/json"

	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/Vupy/cache-toon-queue/src/utils"
)

type Fifo struct{}

func (f *Fifo) AddMessage(queueName string, message structs.ItemOptions) (*structs.QueueItem, error) {
	message = utils.FixMessage(message)

	if message.Priority == 0 {
		return addFifo(structs.RedisQueue, queueName, message)
	}

	return addFifoPriority(structs.RedisQueue, queueName, message)
}

func (f *Fifo) GetMessage(queueName string) (structs.QueueItem, error) {
	var message structs.QueueItem
	item, err := getFifoPriority(structs.RedisQueue, queueName)

	if err != nil || len(item) == 0 {
		itemFifo, err := getFifo(structs.RedisQueue, queueName)

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
