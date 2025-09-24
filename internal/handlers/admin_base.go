package handlers

import (
	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/services"
)

// AdminHandler 管理处理器
type AdminHandler struct {
	service *services.AdminService
	config  *config.ConfigManager
}

func NewAdminHandler(service *services.AdminService, config *config.ConfigManager) *AdminHandler {
	return &AdminHandler{
		service: service,
		config:  config,
	}
}
