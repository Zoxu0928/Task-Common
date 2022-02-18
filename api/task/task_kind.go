package task

import (
	"time"

	"github.com/Masterminds/semver"
)

type TaskKind int

type conf struct {
	// 任务的超时时间
	timeout time.Duration
	// 任务的版本号，创建任务时版本号会写入数据库
	version string
	// 任务接受的语义化版本
	acceptSemVer string
}

// 严禁调整该顺序
const (
	TaskKindUnknown              = -1
	TaskKindVMMigration TaskKind = 1 // 模拟 - 迁移
	// TaskKindVMRestart   TaskKind = 2 // 模拟 - 重启
	// TaskKindVMDetail    TaskKind = 3 // 模拟 - 详情
	// TaskKindVMStop      TaskKind = 4 // 模拟 - 停止

	TaskKindAsyncDemo                  TaskKind = 5  // 异步任务范例
	TaskKindArkReport                  TaskKind = 6  // 大屏 - 组件运行状态
	TaskKindBilling                    TaskKind = 7  // 帐单 - 月度费用更新
	TaskKindTenantDiskResourcesAll     TaskKind = 8  // 大屏 - 云硬盘统计
	TaskKindTenantInstanceResourcesAll TaskKind = 9  // 大屏 - 云主机统计
	TaskKindTenantEipResourcesAll      TaskKind = 10 // 大屏 - EIP机统计
	TaskKindBillingExport			   TaskKind = 11 // 账单 - 账单导出

)

var taskKindText = map[TaskKind]string{
	TaskKindVMMigration: "vm_migration",
	// TaskKindVMRestart:   "vm_restart",
	// TaskKindVMDetail:    "vm_detail",
	// TaskKindVMStop:      "vm_stop",
	TaskKindAsyncDemo:                  "async_demo",
	TaskKindArkReport:                  "ark_report",
	TaskKindBilling:                    "billing_update",
	TaskKindBillingExport: 				"billing_export",
	TaskKindTenantDiskResourcesAll:     "measure/tenant/resources/tenant_disk_resources_all",
	TaskKindTenantInstanceResourcesAll: "measure/tenant/resources/tenant_instance_resources_all",
	TaskKindTenantEipResourcesAll:      "measure/tenant/resources/tenant_eip_resources_all",
}

// 任务配置
var taskKindConf = map[TaskKind]*conf{
	TaskKindVMMigration: {},
	// TaskKindVMRestart:   {},
	// TaskKindVMDetail:    {},
	// TaskKindVMStop:      {},
	TaskKindAsyncDemo: {version: "1.0.0", acceptSemVer: "^1.0.0"},
}

// Timeout 获取任务类型的超时时间，默认0
func (tk TaskKind) Timeout() time.Duration {
	if v, ok := taskKindConf[tk]; ok {
		return v.timeout
	} else {
		return 0
	}
}

// Version 获取任务的版本号，默认空
func (tk TaskKind) Version() string {
	if v, ok := taskKindConf[tk]; ok {
		return v.version
	}
	return ""
}

// AcceptSemVer 获取任务接受的语义化版本，默认万能匹配符
func (tk TaskKind) AcceptSemVer() string {
	if v, ok := taskKindConf[tk]; ok {
		return v.acceptSemVer
	}
	return "*"
}

func (tk TaskKind) String() string {
	if v, ok := taskKindText[tk]; ok {
		return v
	}
	return ""
}

// VersionValidate 校验传入的版本是否通过版本校验（报错也算通过）
// see: https://github.com/Masterminds/semver
func (tk TaskKind) VersionValidate(targetVersion string) bool {
	c, err := semver.NewConstraint(tk.AcceptSemVer())
	if err != nil {
		// Handle constraint not being parseable.
		return true
	}
	v, err := semver.NewVersion(targetVersion)
	if err != nil {
		// Handle version not being parseable.
		return true
	}
	res, _ := c.Validate(v)
	return res
}

func ConvertToTaskKind(obj string) TaskKind {
	for kind, text := range taskKindText {
		if text == obj {
			return kind
		}
	}
	return TaskKindUnknown
}

func GetTaskKindSet() []TaskKind {
	objs := make([]TaskKind, len(taskKindText), len(taskKindText))
	index := 0
	for status := range taskKindText {
		objs[index] = status
		index++
	}
	return objs
}
