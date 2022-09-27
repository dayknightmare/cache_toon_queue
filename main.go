package main

import (
	"github.com/Vupy/cache-toon-queue/src/client"
	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/Vupy/cache-toon-queue/src/subs"
)

func StartQueue(opt structs.Options) {
	structs.OptionsLoaded = &opt
	structs.RedisQueue = structs.NewRedisStruct()
}

func init() {
	structs.Envs.Fill()
}

func main() {
	StartQueue(structs.Options{
		RedisDB:           structs.Envs.RedisDB,
		RedisIsCluster:    structs.Envs.RedisIsCluster,
		RedisHost:         structs.Envs.RedisHost,
		RedisPort:         structs.Envs.RedisPort,
		RedisClusterHosts: structs.Envs.RedisClusterHosts,
	})

	go subs.StartRetrier()
	go subs.StartCleaner()

	client.StartClient()
}
