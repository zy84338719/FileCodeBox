package web

type ControlDataRequest struct {
	Action string `json:"action" binding:"required"` // "start" 或 "stop"
}

type TestDataRequest struct {
	Port string `json:"port"`
	Host string `json:"host"`
}
