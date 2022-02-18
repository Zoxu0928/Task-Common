package task

type TaskStatus int

const (
	TaskStatusUnknown    TaskStatus = -1 // 未知状态
	TaskStatusCreated    TaskStatus = 1  // 任务被成功创建
	TaskStatusDispatched TaskStatus = 2  // 任务被成功分配
	TaskStatusRunning    TaskStatus = 3  // 任务正在被执行
	TaskStatusFailed     TaskStatus = 4  // 任务失败
	TaskStatusSucceed    TaskStatus = 5  // 任务成功
	TaskStatusCanceled   TaskStatus = 6  // 任务被取消，保留字段，暂时没有使用
)

// taskStatusText 状态文本描述
var taskStatusText = map[TaskStatus]string{
	TaskStatusUnknown:    "unknown",
	TaskStatusCreated:    "created",
	TaskStatusDispatched: "dispatched",
	TaskStatusRunning:    "running",
	TaskStatusFailed:     "failed",
	TaskStatusSucceed:    "succeed",
	TaskStatusCanceled:   "canceled",
}

// 状态机流转限制
var statusTransitionsLimit = map[TaskStatus][]TaskStatus{
	TaskStatusCreated:    {TaskStatusDispatched, TaskStatusCanceled},
	TaskStatusDispatched: {TaskStatusRunning},
	TaskStatusRunning:    {TaskStatusFailed, TaskStatusSucceed},
	TaskStatusFailed:     {TaskStatusCanceled},
	TaskStatusSucceed:    {},
	TaskStatusCanceled:   {},
}

// 需要被分配的任务状态
var DispatchedStatus = []TaskStatus{
	TaskStatusCreated,
	TaskStatusDispatched,
	TaskStatusRunning,
}

func (s TaskStatus) String() string {
	if v, ok := taskStatusText[s]; ok {
		return v
	}
	return ""
}

func ConvertToTaskStatus(obj string) TaskStatus {
	for status, text := range taskStatusText {
		if text == obj {
			return status
		}
	}
	return TaskStatusUnknown
}

func GetTaskStatusSet() []TaskStatus {
	objs := make([]TaskStatus, len(taskStatusText), len(taskStatusText))
	index := 0
	for status, _ := range taskStatusText {
		objs[index] = status
		index++
	}
	return objs
}

// 判断是否为终态
func (s TaskStatus) Finished() bool {
	return s == TaskStatusSucceed || s == TaskStatusCanceled
}

// 状态流转校验
func (s TaskStatus) TransitionTo(to TaskStatus) bool {
	if s == 0 || to == 0 {
		return false
	}

	for _, status := range statusTransitionsLimit[s] {
		if status == to {
			return true
		}
	}
	return false
}
