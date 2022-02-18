package task

import "github.com/Zoxu0928/task-common/api"

// DescribeTaskResponse response for describe task
type DescribeTaskResponse struct {
	Task *Task `json:"task"`
}

// DescribeTasksResponse response for describe tasks
type DescribeTasksResponse struct {
	api.Response
	TotalCount int64   `json:"totalCount"`
	Tasks      []*Task `json:"tasks"`
}

// DescribeTasksBriefResponse response for describe tasks brief
type DescribeTasksBriefResponse struct {
	api.Response
	TotalCount int64        `json:"totalCount"`
	Tasks      []*TaskBrief `json:"tasks"`
}

// UpdateTaskResponse response for update task
type UpdateTaskResponse struct {
	api.Response
}
