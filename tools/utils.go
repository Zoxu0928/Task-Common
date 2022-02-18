package tools

import (
	"github.com/Zoxu0928/task-common/logger"
	"regexp"
	"runtime"
	"sync"
	"time"
)

// 等待指定时间后，如果wg没有结束，则返回超时
// true: 超时
// false: 未超时
func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	lock := make(chan struct{})
	go func() {
		defer close(lock)
		wg.Wait()
	}()
	select {
	case <-lock: // wg结束后，此通道会被关闭，进入此case代表未超时
		return false
	case <-time.After(timeout): // 达到超时时间
		return true
	}
}

// 计算分页的起始位置
func SplitIndex(pageNumber, pageSize, total int) (canSplit bool, start int, end int) {
	start = (pageNumber - 1) * pageSize //起始下标，包含
	end = start + pageSize              //结束下标，不包含
	if start < total {
		if end > total {
			end = total
		}
		canSplit = true
	}
	return
}

func TimeTrack(who string, start time.Time) {
	elapsed := time.Since(start).Seconds()
	if who == "" {
		// Skip this function, and fetch the PC and file for its parent.
		pc, _, _, _ := runtime.Caller(1)

		// Retrieve a function object this functions parent.
		funcObj := runtime.FuncForPC(pc)

		// Regex to extract just the function name (and not the module path).
		runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)
		who = runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")
	}
	logger.Debug("%s elapsed %fs", who, elapsed)
}
