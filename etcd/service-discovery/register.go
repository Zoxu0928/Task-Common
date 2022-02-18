package service_discovery

import (
	"context"
	"encoding/json"
	"github.com/Zoxu0928/task-common/logger"
	"strings"
	"time"

	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Register 服务注册
// prefix 表示路径前缀
// ttl 心跳周期，单位是秒
func (s *Service) Register(prefix string, ttl int64) {
	path := strings.TrimRight(prefix, "/") + "/" + s.UUID
	value, _ := json.Marshal(s)

	kv := clientv3.NewKV(s.client)
	lease := clientv3.NewLease(s.client)

	var curLeaseID clientv3.LeaseID = 0

	if ttl <= 0 {
		ttl = 10
	} else if ttl < 3 {
		ttl = 3
	}

	var interval = time.Duration(ttl/3) * time.Second

	for {
		select {
		case <-s.ctx.Done():
			if err := lease.Close(); err != nil {
				logger.Error("[service] [register] failed close lease, %s", err.Error())
			}
			if _, err := kv.Delete(context.TODO(), path); err != nil {
				logger.Error("[service] [register] failed delete %s, %s", path, err.Error())
			}
			return
		default:
		GRANT:
			if curLeaseID == 0 {
				leaseResp, err := lease.Grant(context.TODO(), ttl)
				if err != nil {
					panic(err)
				}

				if _, err := kv.Put(context.TODO(), path, string(value), clientv3.WithLease(leaseResp.ID)); err != nil {
					panic(err)

				}
				curLeaseID = leaseResp.ID
			} else { // 续约租约，如果租约已经过期将curLeaseID复位到0重新走创建租约的逻辑
				if _, err := lease.KeepAliveOnce(context.TODO(), curLeaseID); err == rpctypes.ErrLeaseNotFound {
					curLeaseID = 0
					goto GRANT
				}
			}
		}
		time.Sleep(interval)
	}
}

// Unregister unregister service from etcd
// warning: do not close etcd client because of used by other instances
func (s *Service) Unregister() {
	s.cancel()
}
