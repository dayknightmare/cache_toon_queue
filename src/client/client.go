package client

import (
	usescases "github.com/Vupy/cache-toon-queue/src/client/usesCases"
	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/Vupy/cache-toon-queue/src/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func StartClient() {
	r := gin.Default()
	id_worker := uuid.New().String()

	go structs.HubRoom.Run(id_worker)
	utils.HandlerControllC(id_worker)

	r.GET("/:queue", usescases.Dequeue)
	r.GET("/metrics", usescases.Metrics)
	r.POST("/move_list", usescases.MoveList)
	r.POST("/enqueue", usescases.Enqueue)

	r.Run("0.0.0.0:8555")
}
