package db

import (
	"testing"
	"time"
)

func TestFileCode_IsExpired(t *testing.T) {
	// 当前时间，用于测试
	now := time.Now()

	// 过去的时间（已过期）
	pastTime := now.Add(-24 * time.Hour)

	// 未来的时间（未过期）
	futureTime := now.Add(24 * time.Hour)

	tests := []struct {
		name        string
		fileCode    FileCode
		wantExpired bool
	}{
		{
			name: "未设置过期时间和次数限制",
			fileCode: FileCode{
				ExpiredAt:    nil,
				ExpiredCount: -1, // -1 表示无限制
			},
			wantExpired: false,
		},
		{
			name: "时间已过期",
			fileCode: FileCode{
				ExpiredAt:    &pastTime,
				ExpiredCount: -1, // 无次数限制
			},
			wantExpired: true,
		},
		{
			name: "时间未过期",
			fileCode: FileCode{
				ExpiredAt:    &futureTime,
				ExpiredCount: -1, // 无次数限制
			},
			wantExpired: false,
		},
		{
			name: "次数已用完（为0）",
			fileCode: FileCode{
				ExpiredAt:    nil,
				ExpiredCount: 0,
			},
			wantExpired: true,
		},
		{
			name: "次数未用完（大于0）",
			fileCode: FileCode{
				ExpiredAt:    nil,
				ExpiredCount: 5,
			},
			wantExpired: false,
		},
		{
			name: "次数无限制（小于0）",
			fileCode: FileCode{
				ExpiredAt:    nil,
				ExpiredCount: -1,
			},
			wantExpired: false,
		},
		{
			name: "时间未过期但次数已用完",
			fileCode: FileCode{
				ExpiredAt:    &futureTime,
				ExpiredCount: 0,
			},
			wantExpired: true,
		},
		{
			name: "时间已过期但次数未用完",
			fileCode: FileCode{
				ExpiredAt:    &pastTime,
				ExpiredCount: 5,
			},
			wantExpired: true,
		},
		{
			name: "时间和次数都未过期",
			fileCode: FileCode{
				ExpiredAt:    &futureTime,
				ExpiredCount: 5,
			},
			wantExpired: false,
		},
		{
			name: "时间和次数都已过期",
			fileCode: FileCode{
				ExpiredAt:    &pastTime,
				ExpiredCount: 0,
			},
			wantExpired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fileCode.IsExpired(); got != tt.wantExpired {
				t.Errorf("FileCode.IsExpired() = %v, want %v", got, tt.wantExpired)
			}
		})
	}
}
