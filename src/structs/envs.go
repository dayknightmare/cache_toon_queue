package structs

import (
	"os"
	"strconv"
	"strings"
)

type env struct {
	RedisDB           int
	RedisIsCluster    bool
	RedisHost         string
	RedisClusterHosts []string
	RedisPort         string
	DequeueTrickTime  int
}

var Envs = &env{}

func (e *env) getDefaultValue(key string, default_value string) string {
	value := os.Getenv(key)

	if len(value) == 0 {
		return default_value
	}

	return value
}

func (e *env) Fill() {
	e.RedisIsCluster = strings.ToLower(e.getDefaultValue("REDIS_IS_CLUSTER", "false")) == "true"
	e.RedisHost = e.getDefaultValue("REDIS_HOST", "localhost")
	e.RedisClusterHosts = strings.Split(strings.ReplaceAll(e.getDefaultValue("REDIS_CLUSTER_HOSTS", ""), " ", ""), ",")
	e.RedisPort = e.getDefaultValue("REDIS_PORT", "6379")

	redisDB, err := strconv.Atoi(e.getDefaultValue("REDIS_DB", "0"))

	if err != nil {
		panic("REDIS_DB must be a number")
	}

	e.RedisDB = redisDB

	dequeueTrickTime, err := strconv.Atoi(e.getDefaultValue("DEQUEUE_TRICK_TIME", "150"))

	if err != nil {
		panic("DEQUEUE_TRICK_TIME must be a number in miliseconds")
	}

	e.DequeueTrickTime = dequeueTrickTime
}
