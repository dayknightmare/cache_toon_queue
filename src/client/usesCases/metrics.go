package usescases

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/Vupy/cache-toon-queue/src/structs/models"
	"github.com/gin-gonic/gin"
)

type collectMetricsFuncs func(info map[string]interface{}, sufix string, key string) (map[string]interface{}, error)

type metricsData struct {
	metrics map[string]interface{}
}

func (m *metricsData) mergeOrCreate(queueName string, info map[string]interface{}) {
	if _, ok := m.metrics["queues"].(map[string]interface{})[queueName]; !ok {
		m.metrics["queue_count"] = m.metrics["queue_count"].(int) + 1
		m.metrics["queues"].(map[string]interface{})[queueName] = info
		m.metrics["queues"].(map[string]interface{})[queueName].(map[string]interface{})["workers_count"] = collectWorkersAmount(queueName)
		m.metrics["queues"].(map[string]interface{})[queueName].(map[string]interface{})["workers"] = collectWorkers(queueName)
	} else {
		m.metrics["queues"].(map[string]interface{})[queueName] = mergeValues(
			m.metrics["queues"].(map[string]interface{})[queueName].(map[string]interface{}),
			info,
		)
	}
}

func (m *metricsData) collectQueues(keyPattern string, queuePosition int, statusQueue string) {
	keys, err := structs.RedisQueue.Client.Keys(keyPattern).Result()

	if err != nil {
		return
	}

	funs := map[string]collectMetricsFuncs{
		"completed":  collectCompleted,
		"processing": collectProcessing,
		"waiting":    collectWaitings,
		"failed":     collectFailed,
	}

	for _, key := range keys {
		typeQueue := "fifo"
		keySplited := strings.Split(key, ":")
		sufix := keySplited[queuePosition+1]

		if sufix == "l" || sufix == "zl" {
			typeQueue = "lifo"
		}

		queueName := strings.Replace(keySplited[queuePosition], "cache_toon_", "", 1) + ":" + typeQueue
		info := createInfo(
			strings.Replace(keySplited[queuePosition], "cache_toon_", "", 1),
			typeQueue,
		)

		newInfo, err := funs[statusQueue](info, sufix, key)

		if err != nil {
			continue
		}

		m.mergeOrCreate(
			queueName,
			newInfo,
		)
	}
}

func mergeValues(values map[string]interface{}, additional map[string]interface{}) map[string]interface{} {
	values["waiting"] = values["waiting"].(int) + additional["waiting"].(int)
	values["non_priority"] = values["non_priority"].(int) + additional["non_priority"].(int)
	values["priority"] = values["priority"].(int) + additional["priority"].(int)
	values["completed"] = values["completed"].(int) + additional["completed"].(int)
	values["processing"] = values["processing"].(int) + additional["processing"].(int)
	values["failed"] = values["failed"].(int) + additional["failed"].(int)
	values["workers_count"] = values["workers_count"].(int) + additional["workers_count"].(int)

	return values
}

func createInfo(name string, typeQueue string) map[string]interface{} {
	return map[string]interface{}{
		"name":          name,
		"waiting":       0,
		"non_priority":  0,
		"priority":      0,
		"type_queue":    typeQueue,
		"completed":     0,
		"processing":    0,
		"failed":        0,
		"workers_count": 0,
		"workers":       []string{},
	}
}

func collectWorkersAmount(queueName string) int {
	amount, err := structs.RedisQueue.Client.LLen("wss:" + queueName).Result()

	if err != nil {
		return 0
	}

	return int(amount)
}

func collectWorkers(queueName string) []string {
	workers, err := structs.RedisQueue.Client.LRange("wss:"+queueName, 0, -1).Result()

	if err != nil {
		return []string{}
	}

	return workers
}

func collectWaitings(info map[string]interface{}, sufix string, key string) (map[string]interface{}, error) {
	if sufix == "zf" || sufix == "zl" {
		amount, err := structs.RedisQueue.Client.ZCard(key).Result()

		if err != nil {
			return nil, err
		}

		info["waiting"] = int(amount)
		info["priority"] = int(amount)
	} else {
		amount, err := structs.RedisQueue.Client.LLen(key).Result()

		if err != nil {
			return nil, err
		}

		info["waiting"] = int(amount)
		info["non_priority"] = int(amount)
	}

	return info, nil
}

func collectCompleted(info map[string]interface{}, sufix string, key string) (map[string]interface{}, error) {
	amount, err := structs.RedisQueue.Client.ZCard(key).Result()

	if err != nil {
		return nil, err
	}

	info["completed"] = int(amount)

	return info, nil
}

func collectProcessing(info map[string]interface{}, sufix string, key string) (map[string]interface{}, error) {
	amount, err := structs.RedisQueue.Client.ZCard(key).Result()

	if err != nil {
		return nil, err
	}

	info["processing"] = int(amount)

	return info, nil
}

func collectFailed(info map[string]interface{}, sufix string, key string) (map[string]interface{}, error) {
	amount, err := structs.RedisQueue.Client.ZCard(key).Result()

	if err != nil {
		return nil, err
	}

	info["failed"] = int(amount)

	return info, nil
}

func collectMetrics() map[string]interface{} {
	metrics := map[string]interface{}{}
	metrics["queue_count"] = 0
	metrics["queues"] = map[string]interface{}{}

	m := metricsData{
		metrics: metrics,
	}

	m.collectQueues(
		"cache_toon_*",
		0,
		"waiting",
	)

	m.collectQueues(
		"completed:*",
		1,
		"completed",
	)

	m.collectQueues(
		"processing:*",
		1,
		"processing",
	)

	m.collectQueues(
		"error:*",
		1,
		"failed",
	)

	data, _ := json.Marshal(m.metrics)
	structs.RedisQueue.Client.Set("metrics", string(data), time.Second*30)

	return m.metrics
}

func Metrics(c *gin.Context) {
	// result, err := structs.RedisQueue.Client.Get("metrics").Result()

	// if err == redis.Nil {
	// 	c.JSON(
	// 		http.StatusOK,
	// 		models.ResponseSuccess{
	// 			Data: collectMetrics(),
	// 		},
	// 	)

	// 	return
	// }

	// if err != nil {
	// 	c.JSON(
	// 		http.StatusInternalServerError,
	// 		models.ResponseError{
	// 			Message: "Error getting metrics",
	// 		},
	// 	)

	// 	return
	// }

	// var r interface{}

	// err = json.Unmarshal([]byte(result), &r)

	// if err != nil {
	c.JSON(
		http.StatusOK,
		models.ResponseSuccess{
			Data: collectMetrics(),
		},
	)

	// return
	// }

	// c.JSON(
	// 	http.StatusOK,
	// 	models.ResponseSuccess{
	// 		Data: r,
	// 	},
	// )
}
