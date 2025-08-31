package mcp

import (
	"encoding/json"
	"fmt"
)

// MCP 协议版本
const MCPVersion = "2024-11-05"

// JSONRPCMessage JSON-RPC 2.0 消息
type JSONRPCMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError JSON-RPC 错误
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Standard JSON-RPC error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// InitializeRequest 初始化请求
type InitializeRequest struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ClientCapabilities     `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
	Meta            map[string]interface{} `json:"meta,omitempty"`
}

// ClientCapabilities 客户端能力
type ClientCapabilities struct {
	Roots        *RootsCapability       `json:"roots,omitempty"`
	Sampling     *SamplingCapability    `json:"sampling,omitempty"`
	Experimental map[string]interface{} `json:"experimental,omitempty"`
}

// RootsCapability 根目录能力
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability 采样能力
type SamplingCapability struct{}

// ClientInfo 客户端信息
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult 初始化结果
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

// ServerCapabilities 服务器能力
type ServerCapabilities struct {
	Logging      *LoggingCapability     `json:"logging,omitempty"`
	Prompts      *PromptsCapability     `json:"prompts,omitempty"`
	Resources    *ResourcesCapability   `json:"resources,omitempty"`
	Tools        *ToolsCapability       `json:"tools,omitempty"`
	Experimental map[string]interface{} `json:"experimental,omitempty"`
}

// LoggingCapability 日志能力
type LoggingCapability struct{}

// PromptsCapability 提示能力
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability 资源能力
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ToolsCapability 工具能力
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ServerInfo 服务器信息
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool 工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolsListRequest 工具列表请求
type ToolsListRequest struct {
	Cursor string `json:"cursor,omitempty"`
}

// ToolsListResult 工具列表结果
type ToolsListResult struct {
	Tools      []Tool  `json:"tools"`
	NextCursor *string `json:"nextCursor,omitempty"`
}

// ToolCallRequest 工具调用请求
type ToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ToolCallResult 工具调用结果
type ToolCallResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content 内容
type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data string `json:"data,omitempty"`
	URI  string `json:"uri,omitempty"`
}

// TextContent 文本内容
func TextContent(text string) Content {
	return Content{
		Type: "text",
		Text: text,
	}
}

// Resource 资源定义
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourcesListRequest 资源列表请求
type ResourcesListRequest struct {
	Cursor string `json:"cursor,omitempty"`
}

// ResourcesListResult 资源列表结果
type ResourcesListResult struct {
	Resources  []Resource `json:"resources"`
	NextCursor *string    `json:"nextCursor,omitempty"`
}

// ResourcesReadRequest 资源读取请求
type ResourcesReadRequest struct {
	URI string `json:"uri"`
}

// ResourcesReadResult 资源读取结果
type ResourcesReadResult struct {
	Contents []ResourceContents `json:"contents"`
}

// ResourceContents 资源内容
type ResourceContents struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"` // Base64 encoded
}

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelNotice  LogLevel = "notice"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelCrit    LogLevel = "crit"
	LogLevelAlert   LogLevel = "alert"
	LogLevelEmerg   LogLevel = "emerg"
)

// LoggingMessageNotification 日志消息通知
type LoggingMessageNotification struct {
	Level  LogLevel               `json:"level"`
	Data   interface{}            `json:"data"`
	Logger string                 `json:"logger,omitempty"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

// NewJSONRPCRequest 创建 JSON-RPC 请求
func NewJSONRPCRequest(id interface{}, method string, params interface{}) *JSONRPCMessage {
	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

// NewJSONRPCResponse 创建 JSON-RPC 响应
func NewJSONRPCResponse(id interface{}, result interface{}) *JSONRPCMessage {
	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
}

// NewJSONRPCError 创建 JSON-RPC 错误响应
func NewJSONRPCError(id interface{}, code int, message string, data interface{}) *JSONRPCMessage {
	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

// NewJSONRPCNotification 创建 JSON-RPC 通知
func NewJSONRPCNotification(method string, params interface{}) *JSONRPCMessage {
	return &JSONRPCMessage{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

// Error 实现 error 接口
func (e *RPCError) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// MarshalJSON 自定义 JSON 序列化
func (msg *JSONRPCMessage) MarshalJSON() ([]byte, error) {
	type Alias JSONRPCMessage
	return json.Marshal((*Alias)(msg))
}
