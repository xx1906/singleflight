// in this example, you need to connection to your redis
package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/laxiaohong/singleflight"
	"time"
)

var (
	cli   *redis.Client       = nil
	group *singleflight.Group = singleflight.NewGroup()
)

const (
	clientNum int    = 1024 * 100
	key       string = "hello_world"
	val       string = "你好世界"
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

// read the value from redis without single-flight
func fetchValueFromDataSourceDirect(cli *redis.Client, key string) (value interface{}, err error) {
	return cli.Get(context.TODO(), key).Result()
}
func main() {
	var ch = make(chan int, 100)
	var t = time.Now()
	var doneChan = make(chan struct{})
	for i := 0; i < clientNum; i++ {
		go func(seq int) {
			<-group.DoChan(key, fetchValueFromDataSource(cli, key))
			ch <- seq
		}(i)
	}

	go func(collector int) {
		var cnt int
		for range ch {
			cnt++
			if cnt == collector {
				doneChan <- struct{}{}
				return
			}
		}
	}(clientNum)

	<-doneChan
	fmt.Println(time.Since(t))
	t = time.Now()
	for i := 0; i < clientNum; i++ {
		go func(seq int) {
			_, err := fetchValueFromDataSourceDirect(cli, key)
			if err != nil {
				fmt.Println(err)
			}
			ch <- seq
		}(i)
	}

	go func(collector int) {
		var cnt int
		for range ch {
			cnt++
			if cnt == collector {
				doneChan <- struct{}{}
				return
			}
		}
	}(clientNum)
	<-doneChan
	fmt.Println(time.Since(t))

}
