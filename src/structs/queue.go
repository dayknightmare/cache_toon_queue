package structs

type Queue interface {
	AddMessage(string, ItemOptions) (*QueueItem, error)
	GetMessage(string) (QueueItem, error)
}

type QueueMessage struct {
	Id      string
	Message QueueItem
}
