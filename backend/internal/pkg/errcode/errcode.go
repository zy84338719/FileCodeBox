// Package errcode 提供全站统一业务码定义。
//
// 业务码段位规划：
//   0       成功
//   1xxxx   通用错误（参数/认证/限流/资源不存在）
//   2xxxx   业务错误（分享/取件）
//   3xxxx   存储错误
//   4xxxx   用户/认证错误
//   5xxxx   系统/内部错误
//   9xxxx   第三方错误
//
// 与 internal/pkg/errors 共存：新代码推荐用本包，旧代码保持兼容。
package errcode

// 1xxxx 通用
const (
	CodeSuccess       = 0
	CodeUnknown       = 10000 // 未知错误
	CodeInvalidParam  = 10001 // 参数错误
	CodeUnauthorized  = 10002 // 未登录
	CodeForbidden     = 10003 // 无权限
	CodeNotFound      = 10004 // 资源不存在
	CodeRateLimit     = 10005 // 限流
	CodeUnavailable   = 10006 // 服务不可用
	CodeTimeout       = 10007 // 请求超时
	CodeInternal      = 10008 // 内部错误
	CodeMethodNotAllowed = 10009 // 方法不允许
	CodeTooLarge      = 10010 // 请求体过大
)

// 2xxxx 分享/取件
const (
	CodeShareNotFound      = 20001 // 分享不存在
	CodeShareExpired       = 20002 // 分享过期
	CodeSharePasswordWrong = 20003 // 密码错误
	CodeShareReachLimit    = 20004 // 超过取件次数
	CodePickupCodeNotFound = 20005 // 取件码不存在
	CodePickupCodeExpired  = 20006 // 取件码已过期
	CodePickupCodeExhausted = 20007 // 取件码已用完
	CodeFileNotFound       = 20008 // 文件不存在
	CodeChunkInvalid       = 20009 // 分片无效
)

// 3xxxx 存储
const (
	CodeStorageInit     = 30001 // 存储初始化失败
	CodeStorageQuota    = 30002 // 存储配额超限
	CodeStorageUpload   = 30003 // 上传失败
	CodeStorageDownload = 30004 // 下载失败
	CodeStorageDelete   = 30005 // 删除失败
	CodeStorageScheme   = 30006 // 存储方案不支持
	CodeStorageIO       = 30007 // 存储 IO 错误
	CodePresignFailed   = 30008 // 预签名生成失败
	CodePresignToken    = 30009 // 预签名 token 无效
	CodePresignExpired  = 30010 // 预签名已过期
)

// 4xxxx 用户/认证
const (
	CodeUserNotFound   = 40001 // 用户不存在
	CodeUserExists     = 40002 // 用户已存在
	CodePasswordWrong  = 40003 // 密码错误
	CodeUserDisabled   = 40004 // 用户已禁用
	CodeTokenInvalid   = 40005 // Token 无效
	CodeTokenExpired   = 40006 // Token 过期
	CodeAPIKeyInvalid  = 40007 // API Key 无效
	CodePermissionDeny = 40008 // 权限不足
	CodeRegClosed      = 40009 // 注册已关闭
	CodeEmailInvalid   = 40010 // 邮箱格式错误
	CodePasswordWeak   = 40011 // 密码强度不足
)

// 5xxxx 系统
const (
	CodeDBError    = 50001 // 数据库错误
	CodeCacheError = 50002 // 缓存错误
	CodeConfigErr  = 50003 // 配置错误
	CodeSchedule   = 50004 // 调度错误
	CodeNotInit    = 50005 // 系统未初始化
	CodeAlreadyInit = 50006 // 系统已初始化
)

// 9xxxx 第三方
const (
	CodeThirdPartySMS    = 90001 // 短信发送失败
	CodeThirdPartyEmail  = 90002 // 邮件发送失败
	CodeThirdPartyAPI    = 90003 // 第三方 API 失败
)

// messages 业务码 → 默认文案
var messages = map[int]string{
	CodeSuccess: "success",

	// 1xxxx
	CodeUnknown:         "unknown error",
	CodeInvalidParam:    "invalid parameter",
	CodeUnauthorized:    "unauthorized",
	CodeForbidden:       "forbidden",
	CodeNotFound:        "not found",
	CodeRateLimit:       "too many requests",
	CodeUnavailable:     "service unavailable",
	CodeTimeout:         "request timeout",
	CodeInternal:        "internal server error",
	CodeMethodNotAllowed: "method not allowed",
	CodeTooLarge:        "request too large",

	// 2xxxx
	CodeShareNotFound:       "share not found",
	CodeShareExpired:        "share expired",
	CodeSharePasswordWrong:  "wrong password",
	CodeShareReachLimit:     "reach pickup limit",
	CodePickupCodeNotFound:  "pickup code not found",
	CodePickupCodeExpired:   "pickup code expired",
	CodePickupCodeExhausted: "pickup code exhausted",
	CodeFileNotFound:        "file not found",
	CodeChunkInvalid:        "invalid chunk",

	// 3xxxx
	CodeStorageInit:     "storage init failed",
	CodeStorageQuota:    "storage quota exceeded",
	CodeStorageUpload:   "upload failed",
	CodeStorageDownload: "download failed",
	CodeStorageDelete:   "delete failed",
	CodeStorageScheme:   "storage scheme not supported",
	CodeStorageIO:       "storage io error",
	CodePresignFailed:   "presign failed",
	CodePresignToken:    "presign token invalid",
	CodePresignExpired:  "presign expired",

	// 4xxxx
	CodeUserNotFound:   "user not found",
	CodeUserExists:     "user already exists",
	CodePasswordWrong:  "wrong password",
	CodeUserDisabled:   "user disabled",
	CodeTokenInvalid:   "token invalid",
	CodeTokenExpired:   "token expired",
	CodeAPIKeyInvalid:  "api key invalid",
	CodePermissionDeny: "permission denied",
	CodeRegClosed:      "registration closed",
	CodeEmailInvalid:   "invalid email",
	CodePasswordWeak:   "password too weak",

	// 5xxxx
	CodeDBError:     "database error",
	CodeCacheError:  "cache error",
	CodeConfigErr:   "config error",
	CodeSchedule:    "scheduler error",
	CodeNotInit:     "system not initialized",
	CodeAlreadyInit: "system already initialized",

	// 9xxxx
	CodeThirdPartySMS:   "sms send failed",
	CodeThirdPartyEmail: "email send failed",
	CodeThirdPartyAPI:   "third party api failed",
}

// Message 拿到业务码对应的默认文案
func Message(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return messages[CodeUnknown]
}

// IsSuccess 判断是否成功
func IsSuccess(code int) bool {
	return code == CodeSuccess
}

// IsClientError 判断是否为 4xx 类（参数/资源问题）
func IsClientError(code int) bool {
	return code >= 10001 && code <= 10999
}

// IsServerError 判断是否为 5xx 类（系统/内部问题）
func IsServerError(code int) bool {
	return code >= 50000 && code <= 59999
}
