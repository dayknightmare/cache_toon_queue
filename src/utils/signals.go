package utils

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Vupy/cache-toon-queue/src/structs"
)

func HandlerControllC(id string) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	structs.RedisQueue.Client.SAdd(
		"config:cache_toon_workers",
		id,
	)

	go func() {
		<-c

		structs.RedisQueue.Client.SRem(
			"config:cache_toon_workers",
			id,
		)

		os.Exit(0)
	}()
}
