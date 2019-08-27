package main

import (
	"fmt"
	"github.com/chremoas/services-common/config"
	"github.com/chremoas/services-common/redis"
)

func main() {
	service := fmt.Sprintf("%s.%s.%s", "net.4amlunch.dev", "srv", "perms")
	redisClient := redis.Init("localhost", "", 0, service)
}
