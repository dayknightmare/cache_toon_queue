package usescases

import (
	"net/http"

	"github.com/Vupy/cache-toon-queue/src/app"
	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/Vupy/cache-toon-queue/src/structs/models"
	"github.com/Vupy/cache-toon-queue/src/utils"
	"github.com/gin-gonic/gin"
)

func Enqueue(c *gin.Context) {
	var data models.NewMessageModel

	if should := utils.ShouldBind(c, &data); !should {
		return
	}

	p := 0

	if data.Priority != 0 {
		p = data.Priority
	}

	if data.TypeQueue == "" {
		data.TypeQueue = "fifo"
	}

	msg, err := app.AddMessage(
		data.Queue,
		structs.ItemOptions{
			Priority:  p,
			TypeQueue: data.TypeQueue,
			Attempt:   data.Attempt,
			Value:     data.Value,
		},
	)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			models.ResponseError{
				Message: err,
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		models.ResponseSuccess{
			Data: msg,
		},
	)
}
