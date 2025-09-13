package web

// AuthLoginRequest 登录请求
type AuthLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthLoginResponse 登录响应
type AuthLoginResponse struct {
	Token     string    `json:"token"`
	TokenType string    `json:"token_type"`
	ExpiresIn int64     `json:"expires_in"`
	User      *UserInfo `json:"user"`
}

// AuthRegisterRequest 注册请求
type AuthRegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
}

// AuthRegisterResponse 注册响应
type AuthRegisterResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID            uint   `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Role          string `json:"role"`
	Status        string `json:"status"`
	EmailVerified bool   `json:"email_verified"`
}
