package service

// AuthTokenData 认证令牌数据
type AuthTokenData struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	ExpiresIn int64  `json:"expires_in"`
	UserID    uint   `json:"user_id"`
}

// AuthValidationResult 认证验证结果
type AuthValidationResult struct {
	Valid     bool   `json:"valid"`
	UserID    uint   `json:"user_id"`
	Role      string `json:"role"`
	Username  string `json:"username"`
	ExpiresAt int64  `json:"expires_at"`
}

// PasswordResetData 密码重置数据
type PasswordResetData struct {
	UserID    uint   `json:"user_id"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// SessionData 会话数据
type SessionData struct {
	SessionID string `json:"session_id"`
	UserID    uint   `json:"user_id"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	ExpiresAt int64  `json:"expires_at"`
	IsActive  bool   `json:"is_active"`
}
