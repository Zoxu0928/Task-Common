package tools

import (
	"context"
	"github.com/Zoxu0928/task-common/logger"
	"runtime/debug"
	"time"
)

// 定时任务
type RegularJob struct {
	name   string
	d      time.Duration
	ctx    context.Context
	cancle context.CancelFunc
}

// 创建定时任务
// RegularCall 和 ExpiredCall 不能同时调用
func CreateRegularJob(name string) *RegularJob {
	return &RegularJob{
		name: name,
	}
}

// 设置定时时间
func (tm *RegularJob) SetDuration(d time.Duration) {
	tm.d = d
}

// 停止定时任务
func (tm *RegularJob) Stop() {
	if tm.cancle != nil {
		tm.cancle()
		tm.cancle = nil
		tm.ctx = nil
	}
}

// 是否处于停止状态
func (tm *RegularJob) IsStop() bool {
	return tm.ctx == nil
}

// 定期回调函数，当需要停止定时任务时，直接关闭定时任务即可
func (tm *RegularJob) RegularCall(f func()) {
	if tm.d == time.Duration(0) {
		logger.Error("%s RegularJob failed. duration is 0", tm.name)
		return
	}
	tm.Stop()
	tm.ctx, tm.cancle = context.WithCancel(context.Background())
	var onError = func() {
		if r := recover(); r != nil {
			logger.Error("RegularCall: [%s] 发生未知错误：%s\n%s", tm.name, r, string(debug.Stack()))
		}
	}
	go func(ctx context.Context) {
		defer onError()
		logger.Debug("定时任务启动 %s", tm.name)
		t := time.NewTicker(tm.d)
		defer t.Stop()
		go func() {
			defer onError()
			logger.Debug("RegularCall: %s", tm.name)
			f()
		}()
	f:
		for {
			select {
			case <-ctx.Done():
				break f
			case <-t.C:
				go func() {
					defer onError()
					logger.Debug("RegularCall: %s", tm.name)
					f()
				}()
			}
		}
		logger.Debug("定时任务退出 %s", tm.name)
	}(tm.ctx)
}

// 到时间后回调一次函数后立即退出，只会调用一次，在调用之前，可以随时关闭任务
func (tm *RegularJob) ExpiredCall(f func()) {
	if tm.d == time.Duration(0) {
		logger.Error("%s RegularJob failed. duration is 0", tm.name)
		return
	}
	tm.Stop()
	tm.ctx, tm.cancle = context.WithCancel(context.Background())
	var onError = func() {
		if r := recover(); r != nil {
			logger.Error("ExpiredCall: [%s] 发生未知错误：%s\n%s", tm.name, r, string(debug.Stack()))
		}
	}
	callTime := time.Now().Add(tm.d).Format("2006-01-02 15:04:05")
	go func(ctx context.Context) {
		defer onError()
		logger.Debug("一次性定时任务启动 %s [CallTime=%s]", tm.name, callTime)
		t := time.NewTimer(tm.d)
		defer t.Stop()
		select {
		case <-ctx.Done():
			break
		case <-t.C:
			go func() {
				defer onError()
				logger.Debug("ExpiredCall %s [CallTime=%s]", tm.name, callTime)
				f()
			}()
			tm.Stop()
		}
		logger.Debug("一次性定时任务退出 %s [CallTime=%s]", tm.name, callTime)
	}(tm.ctx)
}
