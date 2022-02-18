package distribute_mutex

import (
	"context"
	"errors"
	"github.com/Zoxu0928/task-common/etcd"
	"github.com/Zoxu0928/task-common/logger"
	"time"

	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	DistributeMutexKind     = "etcd-v3"
	DefaultTTL          int = 30 // 30 seconds
)

var (
	DistributeMutexNilETCDClientError error = errors.New("[distributeMutex] etcd client is nil")
	DistributeMutexNilETCDPrefix      error = errors.New("[distributeMutex] etcd prefix is empty")

	DefaultDistributeMutex *EtcdMutex
)

var (
	_ = IMutex(&EtcdMutex{})
)

// Implementation of distributed lock with Etcd
type EtcdMutex struct {
	prefix    string
	client    *etcd.Client
	session   *concurrency.Session
	etcdMutex *concurrency.Mutex
	ctx       context.Context
	ttl       int
}

// Lock locks the mutex with a cancelable context. If the context is canceled
// while trying to acquire the lock, the mutex tries to clean its stale lock entry.
// Block calling
func (mutex *EtcdMutex) Lock(ctx context.Context) error {
	baseCtx := context.Background()
	if ctx != nil {
		baseCtx = ctx
	}
	return mutex.etcdMutex.Lock(baseCtx)
}

// TryLock locks the mutex if not already locked by another session.
// If lock is held by another session, return immediately after attempting necessary cleanup
// The ctx argument is used for the sending/receiving Txn RPC.
func (mutex *EtcdMutex) TryLock(ctx context.Context) (err error) {
	baseCtx := context.Background()
	if ctx != nil {
		baseCtx = ctx
	}
	// 尝试获取锁
	err = mutex.etcdMutex.TryLock(baseCtx)
	// 获取失败
	if err != nil && err != concurrency.ErrLocked {
		logger.Error("try lock err = %s", err.Error())
		lresp, terr := mutex.client.TimeToLive(baseCtx, mutex.session.Lease())
		if terr != nil {
			logger.Error("time to live err = %s", terr.Error())
		}
		logger.Info("session ttl = %d， session Granted ttl = %d", lresp.TTL, lresp.GrantedTTL)
		// TTL 是租约剩余的 TTL，单位为秒；
		// 租约将在接下来的 TTL + 1 秒之后过期。
		// GrantedTTL 是租约创建/续约时初始授予的时间，单位为秒。
		// keys 是附加到这个租约的 key 的列表。
		if terr != nil || lresp.TTL < 0 {
			// 租约验证出错，或者租约到期
			logger.Info("etcd lease [%x] is expired ! create a new lease......", mutex.session.Lease())
			// 此时错误信息是requested lease not found
			newMutex, distributeErr := NewETCDDistributeMutex(mutex.client, mutex.prefix, time.Duration(mutex.ttl)*time.Second)
			if distributeErr != nil {
				return distributeErr
			}
			mutex.session = newMutex.session
			mutex.etcdMutex = newMutex.etcdMutex
			logger.Info("old lease is closed, open new lease: [%x] ", newMutex.session.Lease())
			// 使用新的session重新获取锁
			err = mutex.TryLock(baseCtx)
		}
	}
	return err
}

func (mutex *EtcdMutex) UnLock(ctx context.Context) error {
	baseCtx := context.Background()
	if ctx != nil {
		baseCtx = ctx
	}
	return mutex.etcdMutex.Unlock(baseCtx)
}

func (mutex *EtcdMutex) Kind() string {
	return DistributeMutexKind
}

func NewETCDDistributeMutex(etcdClient *etcd.Client, prefix string, ttl time.Duration) (*EtcdMutex, error) {
	if etcdClient == nil {
		logger.Error(DistributeMutexNilETCDClientError.Error())
		return nil, DistributeMutexNilETCDClientError
	}

	if prefix == "" {
		logger.Error(DistributeMutexNilETCDPrefix.Error())
		return nil, DistributeMutexNilETCDPrefix
	}

	ttlV := int(ttl.Seconds())
	if ttl <= 0 {
		ttlV = DefaultTTL
	}

	// 通过租约创建session
	session, err := concurrency.NewSession(etcdClient.Client, concurrency.WithTTL(ttlV))
	if err != nil {
		return nil, err
	}

	m := &EtcdMutex{
		prefix:    prefix,
		client:    etcdClient,
		session:   session,
		etcdMutex: concurrency.NewMutex(session, prefix),
		ttl:       ttlV,
	}

	return m, nil
}
