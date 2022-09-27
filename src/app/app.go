package app

import (
	"fmt"
	"strings"

	"github.com/Vupy/cache-toon-queue/src/app/fifo"
	"github.com/Vupy/cache-toon-queue/src/app/lifo"
	"github.com/Vupy/cache-toon-queue/src/structs"
)

func getTypeQueue(typeQueue string) (string, error) {
	alloweds := []string{"fifo", "lifo"}

	if typeQueue == "" {
		return "fifo", nil
	}

	for _, i := range alloweds {
		if strings.ToLower(typeQueue) == i {
			return i, nil
		}
	}

	return "", fmt.Errorf("type queue not allowed")
}

func getQueue(typeQueue string) (structs.Queue, error) {
	typeQueue, err := getTypeQueue(typeQueue)

	if err != nil {
		return nil, err
	}

	if typeQueue == "lifo" {
		return &lifo.Lifo{}, nil
	}

	return &fifo.Fifo{}, nil
}

func AddMessage(queueName string, message structs.ItemOptions) (*structs.QueueItem, error) {
	queue, err := getQueue(message.TypeQueue)

	if err != nil {
		return &structs.QueueItem{}, err
	}

	msg, err := queue.AddMessage(queueName, message)

	structs.RedisQueue.Client.Publish("sub:"+queueName, "1")

	return msg, err
}

func GetMessage(queueName string, typeQueue string) (structs.QueueItem, error) {
	queue, err := getQueue(typeQueue)

	if err != nil {
		return structs.QueueItem{}, err
	}

	return queue.GetMessage(queueName)
}
