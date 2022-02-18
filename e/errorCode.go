package e

type ErrorCode struct {
	Code int
	Type string
}

var (
	CHARGE_OVERDUE               = &ErrorCode{400, "CHARGE_OVERDUE"}
	CHARGE_ARREAR                = &ErrorCode{400, "CHARGE_ARREAR"}
	INVALID_ARGUMENT             = &ErrorCode{400, "INVALID_ARGUMENT"}             // 参数无效
	FAILED_PRECONDITION          = &ErrorCode{400, "FAILED_PRECONDITION"}          // 不满足前置条件
	OUT_OF_RANGE                 = &ErrorCode{400, "OUT_OF_RANGE"}                 // 超出范围
	CONFLICT                     = &ErrorCode{400, "CONFLICT"}                     // 相互冲突
	DUPLICATE                    = &ErrorCode{400, "DUPLICATE"}                    // 重复
	NO_SESSION                   = &ErrorCode{401, "NO_SESSION"}                   // 未登录
	UNAUTHENTICATED              = &ErrorCode{401, "UNAUTHENTICATED"}              // 身份无效
	REAL_NAME_UNAUTHENTICATED    = &ErrorCode{401, "REAL_NAME_UNAUTHENTICATED"}    // 未实名认证
	PERMISSION_DENIED            = &ErrorCode{403, "PERMISSION_DENIED"}            // 没有权限操作资源
	NOT_FOUND                    = &ErrorCode{404, "NOT_FOUND"}                    // 不存在
	ABORTED                      = &ErrorCode{409, "ABORTED"}                      // 无法锁定资源
	ALREADY_EXISTS               = &ErrorCode{409, "ALREADY_EXISTS"}               // 资源已存在
	QUOTA_EXCEEDED               = &ErrorCode{429, "QUOTA_EXCEEDED"}               // 超出配额
	ACCOUNT_BALANCE_INSUFFICIENT = &ErrorCode{429, "ACCOUNT_BALANCE_INSUFFICIENT"} // 用户余额门槛校验未通过
	RATE_LIMIT                   = &ErrorCode{429, "RATE_LIMIT"}                   // 请求过于频繁(交易的错误类型)
	CANCELLED                    = &ErrorCode{499, "CANCELLED"}                    // 请求已被取消
	DATA_LOSS                    = &ErrorCode{500, "DATA_LOSS"}                    // 数据丢失或数据错误
	UNKNOWN                      = &ErrorCode{500, "UNKNOWN"}                      // 未知错误：定位为BUG
	INTERNAL                     = &ErrorCode{500, "INTERNAL"}                     // 系统内错误：定位为BUG
	NOT_IMPLEMENTED              = &ErrorCode{501, "NOT_IMPLEMENTED"}              // 方法未实现
	UNAVAILABLE                  = &ErrorCode{503, "UNAVAILABLE"}                  // 服务不可达，定位为系统已宕机
	DEADLINE_EXCEEDED            = &ErrorCode{504, "DEADLINE_EXCEEDED"}            // 请求过于频繁
)
