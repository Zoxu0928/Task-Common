package main

//
//import (
//	"fmt"
//	"github.com/Zoxu0928/task-common/cache"
//	"github.com/Zoxu0928/task-common/global/env"
//	"github.com/Zoxu0928/task-common/tools"
//	"github.com/garyburd/redigo/redis"
//
//	"time"
//)
//
//func main() {
//
//	env.ConfigPath = ""
//
//	var lc cache.Cache = cache.Local // 本地缓存
//	//var lc cache.Cache = cache.CreateRedisClient(cache.LoadRedisConfig("redis.yaml")) // redis 单节点模式
//	//var lc cache.Cache = cache.CreateRedisCluster(cache.LoadRedisConfig("redis.yaml")) // redis 集群模式
//	//var lc cache.Cache
//	//var err error
//	//lc, err = cache.CreateRedisConnection("redis.yaml") // redis，根据配置生成单节点或集群模式的连接
//
//	//fmt.Println(err)
//
//	for {
//		var cmd, k, v, ex string
//		fmt.Scanln(&cmd, &k, &v, &ex)
//		if cmd == "" {
//			continue
//		}
//
//		switch cmd {
//		case "set":
//			e, _ := tools.ToInt(ex)
//			fmt.Println(lc.Set(k, v, time.Duration(e)*time.Second))
//		case "setnx":
//			e, _ := tools.ToInt(ex)
//			fmt.Println(lc.SetNx(k, v, time.Duration(e)*time.Second))
//		case "get":
//			fmt.Println(redis.String(lc.Get(k)))
//		case "del":
//			fmt.Println(lc.Remove(k))
//		case "time":
//			fmt.Println(lc.Time(k))
//		case "size":
//			fmt.Println(lc.Size())
//		}
//	}
//}
