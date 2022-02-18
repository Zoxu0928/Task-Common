package task

import "github.com/Zoxu0928/task-common/e"

// 任务接口
type TaskService interface {
	// 查询任务详情
	DescribeTask(request *DescribeTaskRequest) (*DescribeTaskResponse, e.ApiError)
	// 查询任务列表
	DescribeTasks(request *DescribeTasksRequest) (*DescribeTasksResponse, e.ApiError)
	// 查询任务列表，返回精简信息
	DescribeTasksBrief(request *DescribeTasksRequest) (*DescribeTasksBriefResponse, e.ApiError)
	// 更新任务
	UpdateTask(request *UpdateTaskRequest) (*UpdateTaskResponse, e.ApiError)
}

type TaskCreator interface {
	// 创建任务
	Create(kind TaskKind, name, creator, description, params string) (string, error)
}
