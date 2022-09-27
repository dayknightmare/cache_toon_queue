package usescases

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Vupy/cache-toon-queue/src/app"
	"github.com/Vupy/cache-toon-queue/src/client/sockets"
	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/Vupy/cache-toon-queue/src/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	skip    = iota
	goAhead = iota
	exit    = iota
)

var upgrader = websocket.Upgrader{
	WriteBufferPool: websocket.DefaultDialer.WriteBufferPool,
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type dequeueData struct {
	queue     string
	id        string
	typeQueue string
	ws        *websocket.Conn
	isClosed  bool
	RedisSub  *redis.PubSub
	dryMode   bool
}

func (d *dequeueData) moveToLock() {
	structs.RedisQueue.Client.SAdd("lock:"+d.queue, d.id)
}

func (d *dequeueData) moveToUnLock() {
	structs.RedisQueue.Client.SRem("lock:"+d.queue, d.id)
}

func (d *dequeueData) moveToProcessing(msg structs.QueueMessage) {
	msgJson, _ := json.Marshal(msg)

	structs.RedisQueue.Client.HSet(
		"map:"+d.queue,
		msg.Id,
		string(msgJson),
	)

	structs.RedisQueue.Client.SAdd(
		"processing:"+utils.Sufixer(d.queue, msg.Message.TypeQueue, msg.Message.Priority),
		msg.Message.Id,
	)
}

func (d *dequeueData) moveToError(msg structs.QueueMessage) {
	out, _ := json.Marshal(msg.Message)
	t := time.Now().UnixMilli()
	z := redis.Z{
		Score:  float64(t),
		Member: out,
	}

	structs.RedisQueue.Client.HDel(
		"map:"+d.queue,
		msg.Id,
	)

	structs.RedisQueue.Client.ZAdd(
		"error:"+utils.Sufixer(d.queue, d.typeQueue, msg.Message.Priority),
		z,
	)
	structs.RedisQueue.Client.SRem(
		"processing:"+utils.Sufixer(d.queue, d.typeQueue, msg.Message.Priority),
		msg.Message.Id,
	)
}

func (d *dequeueData) isLocked() bool {
	result := structs.RedisQueue.Client.SIsMember("lock:"+d.queue, d.id)
	locked, _ := result.Result()

	return locked
}

func (d *dequeueData) deliveryMessage(msg structs.QueueMessage) error {
	msgJson, err := json.Marshal(msg)

	if err != nil {
		return err
	}

	d.moveToLock()
	d.moveToProcessing(msg)
	err = d.ws.WriteMessage(1, msgJson)

	if err != nil {
		d.moveToUnLock()
		d.moveToError(msg)
		return err
	}

	time.Sleep(time.Microsecond * 100)
	return nil
}

func (d *dequeueData) createAndDeliveryMessage(item structs.QueueItem) error {
	msgQ := structs.QueueMessage{
		Id:      d.id,
		Message: item,
	}

	return d.deliveryMessage(msgQ)
}

func (d *dequeueData) getSubMessage(c chan int) {
	for {
		msg, err := d.RedisSub.ReceiveMessage()

		if err != nil {
			continue
		}

		if msg.Payload == "1" {
			c <- 1
			continue
		}

		c <- 0
	}
}

func (d *dequeueData) waitForMessage(c chan int) int {
	if !d.dryMode {
		select {
		case x, ok := <-c:
			if d.isLocked() {
				d.dryMode = true
			}

			if !ok || x != 1 || d.isClosed {
				d.RedisSub.Close()
				d.isClosed = true
				return exit
			}
		default:
			if d.isClosed {
				d.RedisSub.Close()
				return exit
			}

			return skip
		}
	}

	return goAhead
}

func (d *dequeueData) getMessageAndDelivery() int {
	item, err := app.GetMessage(d.queue, d.typeQueue)

	if err != nil && err != redis.Nil {
		d.RedisSub.Close()
		d.isClosed = true
		return exit
	}

	if item.Id == "" {
		d.dryMode = false
		return skip
	}

	if err := d.createAndDeliveryMessage(item); err != nil {
		d.RedisSub.Close()
		return exit
	}

	return goAhead
}

func (d *dequeueData) receiveQueuedMessages() {
	c := make(chan int)

	go d.getSubMessage(c)

	for {
		t := d.waitForMessage(c)

		if t == exit {
			return
		} else if t == skip {
			continue
		}

		t = d.getMessageAndDelivery()

		if t == exit {
			return
		} else if t == skip {
			continue
		}

		if d.dryMode {
			time.Sleep(time.Microsecond * 300)
		}
	}
}

func (d *dequeueData) createSubscription() *structs.Subscription {
	conn := &structs.Connection{
		Send: make(chan []byte, 16384),
		WS:   d.ws,
	}

	d.RedisSub = structs.RedisQueue.Client.Subscribe("sub:" + d.queue)

	sub := &structs.Subscription{
		Conn:      conn,
		Room:      d.queue,
		TypeQueue: d.typeQueue,
		Id:        d.id,
	}

	return sub
}

func (d *dequeueData) closeQueue() {
	d.isClosed = true

	structs.RedisQueue.Client.SRem(
		"lock:"+d.queue,
		d.id,
	)
}

func Dequeue(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		return
	}

	defer func() {
		ws.Close()
	}()

	id := uuid.New().String()
	queue := c.Param("queue")
	typeQueue := c.DefaultQuery("type", "fifo")

	d := dequeueData{
		ws:        ws,
		id:        id,
		queue:     queue,
		typeQueue: typeQueue,
		isClosed:  false,
		dryMode:   true,
	}

	sub := d.createSubscription()

	structs.HubRoom.Register <- *sub

	go sockets.WriteWSMessage(sub)
	go sockets.ReadWSMessage(*sub, d.closeQueue)

	err = ws.WriteMessage(1, []byte(id))

	if err != nil {
		return
	}

	d.receiveQueuedMessages()
}
