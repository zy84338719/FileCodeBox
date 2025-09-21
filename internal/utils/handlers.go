package utils

import (
	"fmt"
	"mime/multipart"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/common"
)

// BindJSONWithValidation 绑定JSON请求并处理验证错误
func BindJSONWithValidation(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return false
	}
	return true
}

// ParseUserIDFromParam 从URL参数解析用户ID
func ParseUserIDFromParam(c *gin.Context, paramName string) (uint, bool) {
	userIDStr := c.Param(paramName)
	userID64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		common.BadRequestResponse(c, "用户ID错误")
		return 0, false
	}
	return uint(userID64), true
}

// ParseIntFromParam 从URL参数解析整数
func ParseIntFromParam(c *gin.Context, paramName string, errorMessage string) (int, bool) {
	paramStr := c.Param(paramName)
	value, err := strconv.Atoi(paramStr)
	if err != nil {
		common.BadRequestResponse(c, errorMessage)
		return 0, false
	}
	return value, true
}

// ParseFileFromForm 从表单解析文件，统一错误处理
func ParseFileFromForm(c *gin.Context, fieldName string) (*multipart.FileHeader, bool) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		common.BadRequestResponse(c, "文件解析失败: "+err.Error())
		return nil, false
	}
	return file, true
}

// ValidateExpireStyle 验证过期样式是否有效
func ValidateExpireStyle(expireStyle string) bool {
	validStyles := []string{"minute", "hour", "day", "week", "month", "year", "forever"}
	for _, style := range validStyles {
		if style == expireStyle {
			return true
		}
	}
	return false
}

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int
	PageSize int
	Search   string
}

// ParsePaginationParams 解析分页参数
func ParsePaginationParams(c *gin.Context) PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}
}

// ExpireParams 过期参数
type ExpireParams struct {
	ExpireValue int
	ExpireStyle string
	RequireAuth bool
}

// ParseExpireParams 解析过期参数（支持POST表单和结构体）
func ParseExpireParams(expireValueStr, expireStyle, requireAuthStr string) (ExpireParams, error) {
	expireValue := 1
	if expireValueStr != "" {
		var err error
		expireValue, err = strconv.Atoi(expireValueStr)
		if err != nil {
			return ExpireParams{}, fmt.Errorf("过期时间参数错误")
		}
	}

	// 验证过期值
	if expireValue < 0 || (expireStyle != "forever" && expireValue == 0) {
		return ExpireParams{}, fmt.Errorf("过期时间参数错误")
	}

	// 默认值处理
	if expireStyle == "" {
		expireStyle = "day"
	}

	// 验证过期样式
	if !ValidateExpireStyle(expireStyle) {
		return ExpireParams{}, fmt.Errorf("过期样式参数错误")
	}

	requireAuth := false
	if requireAuthStr == "true" {
		requireAuth = true
	}

	return ExpireParams{
		ExpireValue: expireValue,
		ExpireStyle: expireStyle,
		RequireAuth: requireAuth,
	}, nil
}

// GetUserIDFromContext 从gin.Context获取用户ID（如果存在）
func GetUserIDFromContext(c *gin.Context) *uint {
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uint)
		return &id
	}
	return nil
}

// GetStringParamRequired 获取必需的字符串参数，如果为空则返回错误
func GetStringParamRequired(c *gin.Context, paramName, errorMessage string) (string, bool) {
	value := c.Param(paramName)
	if value == "" {
		common.BadRequestResponse(c, errorMessage)
		return "", false
	}
	return value, true
}

// GetQueryWithDefault 获取查询参数，如果不存在则返回默认值
func GetQueryWithDefault(c *gin.Context, key, defaultValue string) string {
	if value := c.Query(key); value != "" {
		return value
	}
	return defaultValue
}
