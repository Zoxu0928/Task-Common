// +build go1.9

package golocal

import (
	"sync"
)

// 不要在重要逻辑中使用

var localMap = sync.Map{}

// Parse the goid from runtime.Stack() output. Slow, but it works.
func getGoId() int64 {
	return GetTID()
}

type Local struct {
	RequestId, Pin, UserId, ResType, ResId string
}

func Get() *Local {
	if v, ok := localMap.Load(getGoId()); ok {
		return v.(*Local)
	}
	return nil
}

func Put(value *Local) {
	localMap.LoadOrStore(getGoId(), value)
}

func Remove() {
	localMap.Delete(getGoId())
}

func GetRequestInfo() (string, string, string) {
	if local := Get(); local != nil {
		return local.RequestId, local.Pin, local.UserId
	}
	return "", "", ""
}

func GetInfo() (string, string, string, string, string) {
	if local := Get(); local != nil {
		return local.RequestId, local.Pin, local.UserId, local.ResType, local.ResId
	}
	return "", "", "", "", ""
}
