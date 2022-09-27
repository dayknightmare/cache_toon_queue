package models

type NewMessageModel struct {
	Queue     string `json:"queue" binding:"required"`
	Value     string `json:"value" binding:"required"`
	TypeQueue string `json:"type_queue" binding:"required,oneof=fifo lifo"`
	Attempt   int    `json:"attempt" binding:"required"`
	Priority  int    `json:"priority"`
}
