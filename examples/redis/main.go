// in this example, you need to connection to your redis-server
package main

import (
	"context"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/laxiaohong/singleflight"
)

var (
	cli    *redis.Client       = nil
	group  *singleflight.Group = singleflight.NewGroup()
	engine *gin.Engine         = gin.Default()
)

const (
	key string = "hello_world"
	val string = "你好世界"
)

func init() {
	cli = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	// ...
	if err := cli.Ping(context.TODO()).Err(); err != nil {
		panic(err)
	}
	cli.Set(context.TODO(), key, val, 0)
}

func fetchValueFromDataSource(cli *redis.Client,
	key string) func() (value interface{}, err error) {
	return func() (value interface{}, err error) {
		return cli.Get(context.TODO(), key).Result()
	}
}

func main() {
	pprof.Register(engine)
	engine.GET("/single_flight", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"res": <-group.DoChan(fmt.Sprintf("%s", key), fetchValueFromDataSource(cli, key))})
	})
	if err := engine.Run(":8080"); err != nil {
		_ = fmt.Errorf("error %s", err)
	}
}
