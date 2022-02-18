package task

import (
	"time"
)

// TaskBrief 任务摘要信息
type TaskBrief struct {
	// 记录的唯一标识，不建议采用ID作为标识，因为ID有可能会变
	RefId string `json:"refId"`
	// 任务开始时间
	StartedAt time.Time `json:"startTime"`
	// 任务结束时间
	FinishedAt time.Time `json:"finishTime"`
	// 任务名称
	Name string `json:"name"`
	// 任务类型
	Kind string `json:"kind"`
	// 任务状态
	Status string `json:"status"`
}

// Task 任务详细信息
type Task struct {
	TaskBrief
	// 创建时间
	CreatedAt time.Time `json:"createdAt"`
	// 创建者
	Creator string `json:"creator"`
	// 最后一次更新时间
	UpdatedAt time.Time `json:"updatedAt"`
	// 最后一次更新人
	Updater string `json:"updater" db:"updater"`
	// 任务的执行者
	Owner string `json:"owner"`
	// 任务来源
	SourceCode string `json:"sourceCode"`
	// 任务描述
	Description string `json:"description"`
	// 当前status的简单描述
	Message string `json:"message"`
	// 当前status的详细描述
	Detail string `json:"detail"`
}

// GetBrief 获取任务摘要信息
func (t *Task) GetBrief() *TaskBrief {
	return &t.TaskBrief
}
