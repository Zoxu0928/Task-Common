package e

import (
	"github.com/Zoxu0928/task-common/logger"
	"runtime/debug"
)

func OnError(txt string) {
	if r := recover(); r != nil {
		logger.Error("Got a runtime error %s. %s\n%s", txt, r, string(debug.Stack()))
	}
}
