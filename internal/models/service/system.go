package service

import (
	"runtime"
)

// 编译时通过 -ldflags 传入的全局变量
var (
	// GoVersion 编译使用的 Go 版本
	GoVersion = "unknown"

	// BuildTime 编译时间，格式为 ISO8601
	BuildTime = "unknown"

	// GitCommit Git 提交哈希值
	GitCommit = "unknown"

	// GitBranch Git 分支名称
	GitBranch = "unknown"

	// Version 应用版本号
	Version = "1.10.2"
)

// BuildInfo 构建信息结构体
type BuildInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	GitBranch string `json:"git_branch"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	Arch      string `json:"arch"`
	OS        string `json:"os"`
}

// GetBuildInfo 获取应用构建信息
func GetBuildInfo() *BuildInfo {
	return &BuildInfo{
		Version:   Version,
		GitCommit: GitCommit,
		GitBranch: GitBranch,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(), // 运行时获取真实的Go版本
		Arch:      runtime.GOARCH,
		OS:        runtime.GOOS,
	}
}
