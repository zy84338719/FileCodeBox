package db

import "gorm.io/gorm"

// AdminOperationLog 记录后台运维操作
// Action: 具体操作标识，例如 maintenance.clean_temp
// Target: 操作影响的目标对象描述
// Success: 操作是否成功
// Message: 可读的描述信息
// Actor: 操作者信息
// LatencyMs: 操作耗时（毫秒，可选）
type AdminOperationLog struct {
	gorm.Model
	Action    string `gorm:"size:100;index" json:"action"`
	Target    string `gorm:"size:255" json:"target"`
	Success   bool   `json:"success"`
	Message   string `gorm:"type:text" json:"message"`
	ActorID   *uint  `gorm:"index" json:"actor_id"`
	ActorName string `gorm:"size:100" json:"actor_name"`
	IP        string `gorm:"size:45" json:"ip"`
	LatencyMs int64  `json:"latency_ms"`
}

// AdminOperationLogQuery 查询后台操作日志
// Success: nil 表示忽略； true/false 表示按结果过滤
type AdminOperationLogQuery struct {
	Action   string
	Actor    string
	Success  *bool
	Page     int
	PageSize int
}
