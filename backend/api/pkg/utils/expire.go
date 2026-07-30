package utils

import (
	"errors"
	"strconv"
	"time"
)

// ExpireParams 过期参数
type ExpireParams struct {
	ExpireValue int
	ExpireStyle string
	RequireAuth bool
}

// ParseExpireParams 解析过期参数
func ParseExpireParams(expireValueStr, expireStyle, requireAuthStr string) (*ExpireParams, error) {
	expireValue, err := strconv.Atoi(expireValueStr)
	if err != nil {
		return nil, errors.New("过期值必须是数字")
	}

	if expireValue <= 0 && expireStyle != "forever" {
		return nil, errors.New("过期值必须大于0")
	}

	requireAuth := requireAuthStr == "true"

	return &ExpireParams{
		ExpireValue: expireValue,
		ExpireStyle: expireStyle,
		RequireAuth: requireAuth,
	}, nil
}

// CalculateExpireTime 计算过期时间
func CalculateExpireTime(expireValue int, expireStyle string) *time.Time {
	if expireStyle == "forever" {
		return nil
	}

	// 修复：当不传参数时，设置默认值为1天
	// 这样避免了 expireValue=0 和 expireStyle="" 导致立即过期的问题
	if expireValue <= 0 {
		expireValue = 1
	}
	if expireStyle == "" {
		expireStyle = "day"
	}

	var duration time.Duration
	switch expireStyle {
	case "minute":
		duration = time.Duration(expireValue) * time.Minute
	case "hour":
		duration = time.Duration(expireValue) * time.Hour
	case "day":
		duration = time.Duration(expireValue) * 24 * time.Hour
	case "week":
		duration = time.Duration(expireValue) * 7 * 24 * time.Hour
	case "month":
		duration = time.Duration(expireValue) * 30 * 24 * time.Hour
	case "year":
		duration = time.Duration(expireValue) * 365 * 24 * time.Hour
	default:
		duration = time.Duration(expireValue) * 24 * time.Hour
	}

	expireTime := time.Now().Add(duration)
	return &expireTime
}

// CalculateExpireCount 计算过期次数（-1 表示无限制）
func CalculateExpireCount(expireStyle string, expireValue int) int {
	if expireStyle == "count" {
		return expireValue
	}
	return -1 // 无限制次数
}
