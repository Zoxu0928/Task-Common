package service_discovery

// 服务健康度检查
type ServiceHealthInfo struct {
	TaskTotal   int `json:"task_total"`
	TaskRunning int `json:"running"`
	TaskDone    int `json:"task_done"`

	// 是否已经准备好提供服务
	Ready            bool     `json:"ready"`
	SupportTaskKinds []string `json:"support_task_kinds"`
}

// 主机健康度检查
type HostHealthInfo struct {
	// 增加一些负载信息
}
