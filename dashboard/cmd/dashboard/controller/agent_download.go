package controller

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// Download agent binary
// @Summary Download agent binary
// @Description Download nezha agent binary file
// @Tags public
// @Produce application/octet-stream
// @Success 200 {file} binary
// @Router /nezha-agent [get]
func downloadAgentBinary(c *gin.Context) (any, error) {
	// 构建 agent 文件路径
	agentPath := filepath.Join("data", "agents", "nezha-agent-amd64")

	// 检查文件是否存在
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		return nil, errors.New("agent binary not found")
	}

	// 设置响应头
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=nezha-agent")
	c.Status(200)

	// 发送文件
	c.File(agentPath)

	// 返回 errNoop 表示已处理响应
	return nil, errNoop
}

// Download agent install script
// @Summary Download agent install script
// @Description Download agent installation script
// @Tags public
// @Produce text/plain
// @Success 200 {file} script
// @Router /install.sh [get]
func downloadInstallScript(c *gin.Context) (any, error) {
	// 构建脚本文件路径
	scriptPath := filepath.Join("data", "agents", "install.sh")

	// 检查文件是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, errors.New("install script not found")
	}

	// 设置响应头
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=install.sh")
	c.Status(200)

	// 发送文件
	c.File(scriptPath)

	// 返回 errNoop 表示已处理响应
	return nil, errNoop
}
