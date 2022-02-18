package task

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTaskStatus_TransitionTo(t *testing.T) {
	var status TaskStatus
	passed := status.TransitionTo(TaskStatusCanceled)
	assert.Equal(t, false, passed)

	status = TaskStatus(100)
	passed = status.TransitionTo(TaskStatusCanceled)
	assert.Equal(t, false, passed)
}

func TestTaskStatus_Finished(t *testing.T) {
	var status TaskStatus
	passed := status.Finished()
	assert.Equal(t, false, passed)

	status = TaskStatusCanceled
	passed = status.Finished()
	assert.Equal(t, true, passed)

	status = TaskStatusSucceed
	passed = status.Finished()
	assert.Equal(t, true, passed)

	status = TaskStatusRunning
	passed = status.Finished()
	assert.Equal(t, false, passed)
}
