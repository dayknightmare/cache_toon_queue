package usescases

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/Vupy/cache-toon-queue/src/structs/models"
	"github.com/Vupy/cache-toon-queue/src/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

type moveListData struct {
	workerId string
	queue    string
	id       string
	msg      structs.QueueMessage
}

func (ml *moveListData) removeLock() {
	structs.RedisQueue.Client.SRem("lock:"+ml.queue, ml.workerId)
}

func (ml *moveListData) removeMap() error {
	var q structs.QueueMessage

	msg, err := structs.RedisQueue.Client.HGet(
		"map:"+ml.queue,
		ml.id,
	).Result()

	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(msg), &q)

	if err != nil {
		fmt.Println(err)
		return err
	}

	ml.msg = q

	structs.RedisQueue.Client.HDel(
		"map:"+ml.queue,
		ml.id,
	)

	return nil
}

func (ml *moveListData) moveToCompleted() {
	out, _ := json.Marshal(ml.msg.Message)
	t := time.Now().UnixMilli()
	z := redis.Z{
		Score:  float64(t),
		Member: out,
	}

	structs.RedisQueue.Client.ZAdd(
		"completed:"+utils.Sufixer(ml.queue, ml.msg.Message.TypeQueue, ml.msg.Message.Priority),
		z,
	)
	structs.RedisQueue.Client.SRem(
		"processing:"+utils.Sufixer(ml.queue, ml.msg.Message.TypeQueue, ml.msg.Message.Priority),
		ml.id,
	)
}

func (ml *moveListData) moveToError() {
	out, _ := json.Marshal(ml.msg.Message)
	t := time.Now().UnixMilli()
	z := redis.Z{
		Score:  float64(t),
		Member: out,
	}

	structs.RedisQueue.Client.ZAdd(
		"error:"+utils.Sufixer(ml.queue, ml.msg.Message.TypeQueue, ml.msg.Message.Priority),
		z,
	)
	structs.RedisQueue.Client.SRem(
		"processing:"+utils.Sufixer(ml.queue, ml.msg.Message.TypeQueue, ml.msg.Message.Priority),
		ml.id,
	)
}

func MoveList(c *gin.Context) {
	var data models.MoveListModel

	if should := utils.ShouldBind(c, &data); !should {
		return
	}

	ml := moveListData{
		workerId: data.WorkerId,
		queue:    data.Queue,
		id:       data.Id,
	}

	err := ml.removeMap()
	ml.removeLock()

	if err != nil && err != redis.Nil {
		c.JSON(
			http.StatusBadRequest,
			models.ResponseError{
				Message: "Error on removing message",
			},
		)

		return
	}

	if !data.GotError {
		ml.moveToCompleted()
	} else {
		ml.moveToError()
	}

	c.JSON(
		http.StatusOK,
		models.ResponseSuccess{
			Data: "ok",
		},
	)
}
