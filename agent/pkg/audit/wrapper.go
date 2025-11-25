package audit

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed wrapper.sh
var wrapperScript string

// WrapperManager 包装器管理器
type WrapperManager struct {
	wrapperPath string
}

// NewWrapperManager 创建包装器管理器
func NewWrapperManager() (*WrapperManager, error) {
	// 创建临时目录
	tmpDir := os.TempDir()
	wrapperPath := filepath.Join(tmpDir, fmt.Sprintf("nezha-audit-wrapper-%d.sh", os.Getpid()))

	// 写入包装器脚本
	if err := os.WriteFile(wrapperPath, []byte(wrapperScript), 0700); err != nil {
		return nil, fmt.Errorf("write wrapper script: %w", err)
	}

	return &WrapperManager{
		wrapperPath: wrapperPath,
	}, nil
}

// GetWrapperPath 获取包装器路径
func (w *WrapperManager) GetWrapperPath() string {
	return w.wrapperPath
}

// Cleanup 清理包装器文件
func (w *WrapperManager) Cleanup() error {
	if w.wrapperPath != "" {
		return os.Remove(w.wrapperPath)
	}
	return nil
}

// DetectShell 检测系统默认 Shell
func DetectShell() string {
	// 优先使用环境变量
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}

	// 尝试常见的 Shell
	shells := []string{
		"/bin/bash",
		"/bin/zsh",
		"/bin/fish",
		"/bin/sh",
	}

	for _, shell := range shells {
		if _, err := os.Stat(shell); err == nil {
			return shell
		}
	}

	// 默认返回 sh
	return "/bin/sh"
}
