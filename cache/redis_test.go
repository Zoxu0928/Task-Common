package cache

//
//import (
//	"fmt"
//	"github.com/Zoxu0928/task-common/global/env"
//	"github.com/garyburd/redigo/redis"
//
//	"testing"
//	"time"
//)
//
//var lc Cache
//
//func TestRedis(t *testing.T) {
//
//	env.ConfigPath = ""
//
//	//var lc cache.Cache = cache.CrateLocalCache() // 本地缓存
//	//var lc cache.Cache = cache.CreateRedisClient(cache.LoadRedisConfig("redis.yaml")) // redis 单节点模式
//	//var lc cache.Cache = cache.CreateRedisCluster(cache.LoadRedisConfig("redis.yaml")) // redis 集群模式
//
//	var err error
//	lc, err = CreateRedisConnection("redis.yaml") // redis，根据配置生成单节点或集群模式的连接
//
//	fmt.Println(err)
//
//	for i := 1; i <= 9; i++ {
//		go get(fmt.Sprintf("key%d", i))
//	}
//
//	time.Sleep(time.Minute)
//}
//
//func get(key string) {
//	fmt.Println(key)
//	for {
//		f := time.Now()
//		val, err := redis.String(lc.Get(key))
//		fmt.Println(key, val, "err:", err, "pool:", lc.Mon(), time.Since(f))
//		time.Sleep(11 * time.Millisecond)
//	}
//}
