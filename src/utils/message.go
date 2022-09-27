package utils

import (
	"github.com/Vupy/cache-toon-queue/src/structs"
)

func limitPriority(message structs.ItemOptions) structs.ItemOptions {
	if message.Priority > 100 {
		message.Priority = 100
	}

	if message.Priority < 0 {
		message.Priority = 0
	}

	return message
}

func limitAttempt(message structs.ItemOptions) structs.ItemOptions {
	if message.Priority > 100 {
		message.Priority = 100
	}

	if message.Attempt < 0 {
		message.Attempt = 0
	}

	return message
}

func FixMessage(message structs.ItemOptions) structs.ItemOptions {
	message = limitAttempt(message)
	message = limitPriority(message)

	return message
}
