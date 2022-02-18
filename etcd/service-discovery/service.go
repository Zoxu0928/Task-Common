package service_discovery

import (
	"context"
	"github.com/Zoxu0928/task-common/api/task"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type IService interface {
}

var _ = IService(&Service{})

// Urls is a representation of the health check urls.
// The key is the protocol name and the value is the url for this protocol.
// Typical usage is:
// 	Ports{
//		"http":"http://127.0.0.1?Action=ServiceHealthCheck",
//		"https": "https://127.0.0.1?Action=ServiceHealthCheck",
//  }
type Urls map[string]string

// Service store all the host information
type Service struct {
	// 主机名称
	HostName string `json:"host_name"`
	// 主机IP
	IP string `json:"ip"`

	HealthCheckUrls Urls   `json:"health_check_urls"`
	UUID            string `json:"uuid"`

	// 支持的任务类型
	SupportTaskKinds []task.TaskKind `json:"support_task_kinds"`

	ctx    context.Context
	cancel context.CancelFunc
	client *clientv3.Client
}

func NewService(client *clientv3.Client, taskKinds []task.TaskKind, hostname, ip, uuid string, urls map[string]string) *Service {
	s := &Service{
		HostName:         hostname,
		IP:               ip,
		HealthCheckUrls:  Urls(urls),
		UUID:             uuid,
		SupportTaskKinds: taskKinds,
		client:           client,
	}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}

// todo 待实现
func (s *Service) HealthCheck() error {
	if url := s.URL("http"); url != "" {
		return nil
	}
	return nil
}

func (s *Service) URL(scheme string) string {
	if s.IP == "" {
		return ""
	}
	url, ok := s.HealthCheckUrls[scheme]
	if !ok {
		return ""
	}
	return url
}
