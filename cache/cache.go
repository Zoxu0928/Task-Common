package cache

import "time"

// 接口
type Cache interface {
	// 增加缓存
	Set(k string, v interface{}, period time.Duration) error

	// 仅当键值不存在时设置键值，并返回1
	// 当键值存在时，返回-1，
	SetNx(k string, v interface{}, period time.Duration) (int, error)

	// 获取缓存
	Get(k string) (interface{}, error)

	// 删除缓存
	Remove(k string) error

	// 获得key的剩余时间秒数
	Time(k string) (int, error)

	// 缓存大小
	Size() (int, error)

	// 关闭连接
	Close()

	// 监控
	Mon() string

	// 筛选key
	Keys(key string) ([]string, error)
}
