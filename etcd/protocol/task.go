package protocol

import (
	"github.com/Zoxu0928/task-common/api/task"
	"time"
)

const (
	ServiceRegisterPath = "/services/pcd/middlewares"
	TaskSubscribePath   = "/tasks/pcd/middlewares"
)

type Task struct {
	RefID     string        `json:"ref_id" validate:"required"`
	Owner     string        `json:"owner" validate:"required"`
	Kind      task.TaskKind `json:"kind" validate:"required"`
	CreatedAt time.Time     `json:"created_at" validate:"required"`
	Creator   string        `json:"creator"`
}
