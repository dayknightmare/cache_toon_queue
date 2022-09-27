package structs

import (
	"time"

	"github.com/google/uuid"
)

type ItemOptions struct {
	Value     string `required:"true"`
	Attempt   int    `required:"true"`
	TypeQueue string
	Priority  int
}

type QueueItem struct {
	Id        string
	Value     string
	TypeQueue string
	Priority  int
	CreatedAt time.Time
	UpdatedAt time.Time
	Attempt   int
}

func NewQueueItem(item ItemOptions) *QueueItem {
	return &QueueItem{
		Id:        uuid.New().String(),
		Value:     item.Value,
		Priority:  item.Priority,
		TypeQueue: item.TypeQueue,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Attempt:   item.Attempt,
	}
}
