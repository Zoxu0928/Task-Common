package task

import (
	"github.com/Zoxu0928/task-common/api"
	"github.com/Zoxu0928/task-common/db"
)

type DescribeTaskRequest struct {
	api.Request
	RefID string `json:"refId"`
}

type DescribeTasksRequest struct {
	api.Request
	db.Pages

	Filters []*api.Filter `json:"filters"`
	Tags    []*api.TagFilter
	// 下面的参数是经过网关鉴权后，表示用户有权限查看的条件
	// 优先使用鉴权后的条件，如果下面参数不为空，说明用户能查看的资源是受限的
	FilterGroups []*api.FilterGroup `json:"filterGroups"`
}

type UpdateTaskRequest struct {
	api.Request
	RefID       string `json:"refId"`
	Owner       string `json:"owner"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	Detail      string `json:"detail"`
	Description string `json:"description"`
}
