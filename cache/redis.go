package cache

//
//// redis 缓存
//
//import (
//	"errors"
//	"fmt"
//	"github.com/Zoxu0928/task-common/logger"
//	"github.com/Zoxu0928/task-common/tools"
//	"github.com/garyburd/redigo/redis"
//	redisCluster "github.com/redis-go-cluster-master"
//	"strings"
//	"sync"
//	"time"
//)
//
//const (
//	REDIS_SET    = "SET"
//	REDIS_GET    = "GET"
//	REDIS_SECOND = "EX"     //过期时间（秒）
//	REDIS_NX     = "NX"     //仅当键值不存在时设置键值
//	REDIS_DEL    = "DEL"    //删除
//	REDIS_TTL    = "TTL"    //剩余时间
//	REDIS_DBSIZE = "DBSIZE" //库大小
//	REDIS_KEYS   = "KEYS"   //筛选key
//)
//
//// Redis客户端
//type redisCache struct {
//	lock sync.Mutex
//
//	// 集群模式连接，内部每个Node维护一个pool
//	cluster *redisCluster.Cluster
//
//	// 单点模式连接
//	pool *redis.Pool
//}
//
//// redis配置
//type RedisConfig struct {
//	Url          string        `yaml:"redis.url"`
//	Auth         string        `yaml:"redis.auth"`
//	ConnTimeout  time.Duration `yaml:"redis.connTimeout"`
//	ReadTimeout  time.Duration `yaml:"redis.readTimeout"`
//	WriteTimeout time.Duration `yaml:"redis.writeTimeout"`
//	AliveTime    time.Duration `yaml:"redis.aliveTime"` //连接存活时间/空闲连接超时时间
//	MaxActive    int           `yaml:"redis.maxActive"` //最大连接数
//	Cluster      bool          `yaml:"redis.cluster"`   //是否集群模式
//	Wait         bool          `yaml:"redis.wait"`      //池中无可用连接时是否等待，只有单点模式pool有意义，集群模式当pool中无连接时会产生新连接
//}
//
//// 加载配置
//func LoadRedisConfig(config string) *RedisConfig {
//	cfg := &RedisConfig{}
//	if err := tools.LoadYaml(config, cfg); err != nil {
//		panic(err)
//	}
//	return cfg
//}
//
//// 实例化，根据配置中的cluster标识，判断注册为单点，或集群模式
//// 消化掉内部panic，即使连接失败，也会返回实例对象
//func CreateRedisConnection(configName string) (conn *redisCache, err error) {
//	cfg := LoadRedisConfig(configName)
//	if cfg.Cluster {
//		conn, err = createRedisCluster(cfg)
//	} else {
//		conn, err = createRedisPool(cfg)
//	}
//	if err != nil {
//		logger.Error("Create redis connection error. %s", err)
//	}
//	return
//}
//
//// 实例化单点连接
//func createRedisPool(cfg *RedisConfig) (*redisCache, error) {
//
//	logger.Info(" *** redis config *** %s", tools.ToJson(cfg))
//
//	// 连接选项
//	op1 := redis.DialReadTimeout(cfg.ReadTimeout)
//	op2 := redis.DialWriteTimeout(cfg.WriteTimeout)
//	op3 := redis.DialConnectTimeout(cfg.ConnTimeout)
//	var op4 redis.DialOption
//	if cfg.Auth != "" {
//		op4 = redis.DialPassword(cfg.Auth)
//	}
//
//	var options []redis.DialOption
//	if cfg.Auth != "" {
//		options = []redis.DialOption{op1, op2, op3, op4}
//	} else {
//		options = []redis.DialOption{op1, op2, op3}
//	}
//
//	// 连接池
//	pool := &redis.Pool{
//		Dial: func() (redis.Conn, error) {
//			return redis.Dial("tcp", cfg.Url, options...)
//		},
//		MaxActive:   cfg.MaxActive,
//		IdleTimeout: cfg.AliveTime,
//		Wait:        cfg.Wait,
//		MaxIdle:     8,
//		TestOnBorrow: func(c redis.Conn, t time.Time) error {
//			_, err := c.Do("PING")
//			return err
//		},
//	}
//
//	// 校验连接是否可用
//	conn := pool.Get()
//	defer conn.Close()
//
//	return &redisCache{pool: pool}, conn.Err()
//}
//
//// 实例化集群连接
//func createRedisCluster(cfg *RedisConfig) (*redisCache, error) {
//
//	logger.Info(" *** redis cluster config *** %s", tools.ToJson(cfg))
//
//	cluster, err := redisCluster.NewCluster(
//		&redisCluster.Options{
//			StartNodes:   strings.Split(cfg.Url, ","),
//			Password:     cfg.Auth,
//			ConnTimeout:  cfg.ConnTimeout,
//			ReadTimeout:  cfg.ReadTimeout,
//			WriteTimeout: cfg.WriteTimeout,
//			AliveTime:    cfg.AliveTime,
//			KeepAlive:    cfg.MaxActive,
//		})
//
//	return &redisCache{cluster: cluster}, err
//}
//
//func (this *redisCache) Set(k string, v interface{}, period time.Duration) error {
//	if period > 0 {
//		stat, err := this.do(REDIS_SET, k, v, REDIS_SECOND, int(period/time.Second))
//		if err != nil {
//			return err
//		}
//		if stat != "OK" {
//			return errors.New(fmt.Sprintf("%s写入缓存失败", k))
//		}
//	} else {
//		stat, err := this.do(REDIS_SET, k, v)
//		if err != nil {
//			return err
//		}
//		if stat != "OK" {
//			return errors.New(fmt.Sprintf("%s写入缓存失败", k))
//		}
//	}
//	return nil
//}
//
//func (this *redisCache) SetNx(k string, v interface{}, period time.Duration) (int, error) {
//	if period > 0 {
//		stat, err := this.do(REDIS_SET, k, v, REDIS_SECOND, int(period/time.Second), REDIS_NX)
//		if err != nil {
//			return -1, err
//		}
//		if stat == nil && err == nil {
//			return -1, nil
//		}
//		if stat == "OK" {
//			return 1, nil
//		}
//	} else {
//		stat, err := this.do(REDIS_SET, k, v, REDIS_NX)
//		if err != nil {
//			return -1, err
//		}
//		if stat == nil && err == nil {
//			return -1, nil
//		}
//		if stat == "OK" {
//			return 1, nil
//		}
//	}
//	return -1, nil
//}
//
//func (this *redisCache) Get(k string) (interface{}, error) {
//	val, err := this.do(REDIS_GET, k)
//	if err != nil {
//		return nil, err
//	} else {
//		return val, nil
//	}
//}
//
//func (this *redisCache) Remove(k string) error {
//	_, err := this.do(REDIS_DEL, k)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (this *redisCache) Time(k string) (int, error) {
//	stat, err := this.do(REDIS_TTL, k)
//	if err != nil {
//		return -1, err
//	}
//	return int(stat.(int64)), nil
//}
//
//// 集群模式只能获取当前连接到的节点的dbsize，不能获取全部节点的总和
//func (this *redisCache) Size() (int, error) {
//	if this.pool != nil {
//		conn := this.pool.Get()
//		defer conn.Close()
//		stat, err := conn.Do(REDIS_DBSIZE)
//		if err != nil {
//			return -1, err
//		}
//		return int(stat.(int64)), nil
//	}
//	return -1, nil
//}
//
//// 执行redis命令
//func (this *redisCache) do(cmd string, args ...interface{}) (interface{}, error) {
//	if this.cluster != nil {
//		return this.cluster.Do(cmd, args...)
//	} else if this.pool != nil {
//		conn := this.pool.Get()
//		defer conn.Close()
//		return conn.Do(cmd, args...)
//	} else {
//		return nil, errors.New("redis连接没有就绪")
//	}
//}
//
//// 关闭连接
//func (this *redisCache) Close() {
//	logger.Info("[Application close] redis client shutting down.")
//	if this.cluster != nil {
//		this.cluster.Close()
//	}
//	if this.pool != nil {
//		this.pool.Close()
//	}
//}
//
//func (this *redisCache) Mon() string {
//	return fmt.Sprintf("%t %d %d %d", this.pool.Wait, this.pool.MaxActive, this.pool.ActiveCount(), this.pool.IdleCount())
//}
//
//// 筛选key
//func (this *redisCache) Keys(key string) ([]string, error) {
//	return redis.Strings(this.do(REDIS_KEYS, key))
//}
