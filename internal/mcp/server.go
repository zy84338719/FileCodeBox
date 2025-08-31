package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// Server MCP 服务器
type Server struct {
	info            ServerInfo
	capabilities    ServerCapabilities
	tools           map[string]Tool
	toolHandlers    map[string]ToolHandler
	resources       map[string]Resource
	resourceReaders map[string]ResourceReader

	// 连接管理
	connections map[string]*Connection
	connMutex   sync.RWMutex

	// 服务器状态
	initialized bool
	initMutex   sync.RWMutex

	// 配置
	instructions string

	// 日志
	logger *logrus.Logger
}

// Connection 连接对象
type Connection struct {
	ID       string
	Reader   *bufio.Scanner
	Writer   io.Writer
	Context  context.Context
	Cancel   context.CancelFunc
	Metadata map[string]interface{}
}

// ToolHandler 工具处理器函数
type ToolHandler func(ctx context.Context, arguments map[string]interface{}) (*ToolCallResult, error)

// ResourceReader 资源读取器函数
type ResourceReader func(ctx context.Context, uri string) (*ResourcesReadResult, error)

// NewServer 创建新的 MCP 服务器
func NewServer(name, version string) *Server {
	return &Server{
		info: ServerInfo{
			Name:    name,
			Version: version,
		},
		capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
			Resources: &ResourcesCapability{
				Subscribe:   false,
				ListChanged: false,
			},
			Logging: &LoggingCapability{},
		},
		tools:           make(map[string]Tool),
		toolHandlers:    make(map[string]ToolHandler),
		resources:       make(map[string]Resource),
		resourceReaders: make(map[string]ResourceReader),
		connections:     make(map[string]*Connection),
		logger:          logrus.New(),
	}
}

// SetInstructions 设置服务器说明
func (s *Server) SetInstructions(instructions string) {
	s.instructions = instructions
}

// AddTool 添加工具
func (s *Server) AddTool(tool Tool, handler ToolHandler) {
	s.tools[tool.Name] = tool
	s.toolHandlers[tool.Name] = handler
}

// AddResource 添加资源
func (s *Server) AddResource(resource Resource, reader ResourceReader) {
	s.resources[resource.URI] = resource
	s.resourceReaders[resource.URI] = reader
}

// ServeStdio 通过 stdio 提供服务
func (s *Server) ServeStdio() error {
	s.logger.Info("Starting MCP server on stdio")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn := &Connection{
		ID:       "stdio",
		Reader:   bufio.NewScanner(os.Stdin),
		Writer:   os.Stdout,
		Context:  ctx,
		Cancel:   cancel,
		Metadata: make(map[string]interface{}),
	}

	s.connMutex.Lock()
	s.connections["stdio"] = conn
	s.connMutex.Unlock()

	defer func() {
		s.connMutex.Lock()
		delete(s.connections, "stdio")
		s.connMutex.Unlock()
	}()

	return s.handleConnection(conn)
}

// ServeTCP 通过 TCP 提供服务
func (s *Server) ServeTCP(addr string) error {
	s.logger.Infof("Starting MCP server on TCP %s", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Errorf("Failed to accept connection: %v", err)
			continue
		}

		go s.handleTCPConnection(conn)
	}
}

// handleTCPConnection 处理 TCP 连接
func (s *Server) handleTCPConnection(tcpConn net.Conn) {
	defer tcpConn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connID := fmt.Sprintf("tcp-%s", tcpConn.RemoteAddr().String())
	conn := &Connection{
		ID:       connID,
		Reader:   bufio.NewScanner(tcpConn),
		Writer:   tcpConn,
		Context:  ctx,
		Cancel:   cancel,
		Metadata: make(map[string]interface{}),
	}

	s.connMutex.Lock()
	s.connections[connID] = conn
	s.connMutex.Unlock()

	defer func() {
		s.connMutex.Lock()
		delete(s.connections, connID)
		s.connMutex.Unlock()
	}()

	s.logger.Infof("New TCP connection: %s", connID)

	if err := s.handleConnection(conn); err != nil {
		s.logger.Errorf("Connection error for %s: %v", connID, err)
	}

	s.logger.Infof("TCP connection closed: %s", connID)
}

// handleConnection 处理连接
func (s *Server) handleConnection(conn *Connection) error {
	for conn.Reader.Scan() {
		select {
		case <-conn.Context.Done():
			return conn.Context.Err()
		default:
		}

		line := strings.TrimSpace(conn.Reader.Text())
		if line == "" {
			continue
		}

		var msg JSONRPCMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			s.logger.Errorf("Failed to parse JSON-RPC message: %v", err)
			s.sendError(conn, nil, ParseError, "Parse error", nil)
			continue
		}

		if err := s.handleMessage(conn, &msg); err != nil {
			s.logger.Errorf("Failed to handle message: %v", err)
		}
	}

	if err := conn.Reader.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

// handleMessage 处理消息
func (s *Server) handleMessage(conn *Connection, msg *JSONRPCMessage) error {
	// 处理请求
	if msg.Method != "" && msg.ID != nil {
		return s.handleRequest(conn, msg)
	}

	// 处理通知
	if msg.Method != "" && msg.ID == nil {
		return s.handleNotification(conn, msg)
	}

	// 忽略响应消息（客户端发送的响应）
	return nil
}

// handleRequest 处理请求
func (s *Server) handleRequest(conn *Connection, msg *JSONRPCMessage) error {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(conn, msg)
	case "tools/list":
		return s.handleToolsList(conn, msg)
	case "tools/call":
		return s.handleToolsCall(conn, msg)
	case "resources/list":
		return s.handleResourcesList(conn, msg)
	case "resources/read":
		return s.handleResourcesRead(conn, msg)
	default:
		return s.sendError(conn, msg.ID, MethodNotFound, "Method not found", nil)
	}
}

// handleNotification 处理通知
func (s *Server) handleNotification(conn *Connection, msg *JSONRPCMessage) error {
	switch msg.Method {
	case "initialized":
		s.initMutex.Lock()
		s.initialized = true
		s.initMutex.Unlock()
		s.logger.Info("Client initialized")
		return nil
	case "notifications/cancelled":
		// 处理取消通知
		s.logger.Debug("Received cancellation notification")
		return nil
	default:
		s.logger.Warnf("Unknown notification method: %s", msg.Method)
		return nil
	}
}

// handleInitialize 处理初始化
func (s *Server) handleInitialize(conn *Connection, msg *JSONRPCMessage) error {
	var req InitializeRequest
	if err := s.parseParams(msg.Params, &req); err != nil {
		return s.sendError(conn, msg.ID, InvalidParams, "Invalid parameters", err.Error())
	}

	s.logger.Infof("Initialize request from %s %s", req.ClientInfo.Name, req.ClientInfo.Version)

	result := InitializeResult{
		ProtocolVersion: MCPVersion,
		Capabilities:    s.capabilities,
		ServerInfo:      s.info,
		Instructions:    s.instructions,
	}

	return s.sendResponse(conn, msg.ID, result)
}

// handleToolsList 处理工具列表
func (s *Server) handleToolsList(conn *Connection, msg *JSONRPCMessage) error {
	tools := make([]Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}

	result := ToolsListResult{
		Tools: tools,
	}

	return s.sendResponse(conn, msg.ID, result)
}

// handleToolsCall 处理工具调用
func (s *Server) handleToolsCall(conn *Connection, msg *JSONRPCMessage) error {
	var req ToolCallRequest
	if err := s.parseParams(msg.Params, &req); err != nil {
		return s.sendError(conn, msg.ID, InvalidParams, "Invalid parameters", err.Error())
	}

	handler, exists := s.toolHandlers[req.Name]
	if !exists {
		return s.sendError(conn, msg.ID, MethodNotFound, "Tool not found", req.Name)
	}

	result, err := handler(conn.Context, req.Arguments)
	if err != nil {
		return s.sendError(conn, msg.ID, InternalError, "Tool execution error", err.Error())
	}

	return s.sendResponse(conn, msg.ID, result)
}

// handleResourcesList 处理资源列表
func (s *Server) handleResourcesList(conn *Connection, msg *JSONRPCMessage) error {
	resources := make([]Resource, 0, len(s.resources))
	for _, resource := range s.resources {
		resources = append(resources, resource)
	}

	result := ResourcesListResult{
		Resources: resources,
	}

	return s.sendResponse(conn, msg.ID, result)
}

// handleResourcesRead 处理资源读取
func (s *Server) handleResourcesRead(conn *Connection, msg *JSONRPCMessage) error {
	var req ResourcesReadRequest
	if err := s.parseParams(msg.Params, &req); err != nil {
		return s.sendError(conn, msg.ID, InvalidParams, "Invalid parameters", err.Error())
	}

	reader, exists := s.resourceReaders[req.URI]
	if !exists {
		return s.sendError(conn, msg.ID, MethodNotFound, "Resource not found", req.URI)
	}

	result, err := reader(conn.Context, req.URI)
	if err != nil {
		return s.sendError(conn, msg.ID, InternalError, "Resource read error", err.Error())
	}

	return s.sendResponse(conn, msg.ID, result)
}

// sendResponse 发送响应
func (s *Server) sendResponse(conn *Connection, id interface{}, result interface{}) error {
	response := NewJSONRPCResponse(id, result)
	return s.sendMessage(conn, response)
}

// sendError 发送错误
func (s *Server) sendError(conn *Connection, id interface{}, code int, message string, data interface{}) error {
	response := NewJSONRPCError(id, code, message, data)
	return s.sendMessage(conn, response)
}

// sendNotification 发送通知
func (s *Server) sendNotification(conn *Connection, method string, params interface{}) error {
	notification := NewJSONRPCNotification(method, params)
	return s.sendMessage(conn, notification)
}

// sendMessage 发送消息
func (s *Server) sendMessage(conn *Connection, msg *JSONRPCMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	data = append(data, '\n')

	_, err = conn.Writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// parseParams 解析参数
func (s *Server) parseParams(params interface{}, target interface{}) error {
	if params == nil {
		return nil
	}

	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

// LogMessage 发送日志消息到客户端
func (s *Server) LogMessage(level LogLevel, data interface{}, logger string) {
	notification := LoggingMessageNotification{
		Level:  level,
		Data:   data,
		Logger: logger,
	}

	s.connMutex.RLock()
	defer s.connMutex.RUnlock()

	for _, conn := range s.connections {
		if err := s.sendNotification(conn, "notifications/message", notification); err != nil {
			log.Printf("Failed to send log notification: %v", err)
		}
	}
}

// IsInitialized 检查是否已初始化
func (s *Server) IsInitialized() bool {
	s.initMutex.RLock()
	defer s.initMutex.RUnlock()
	return s.initialized
}
