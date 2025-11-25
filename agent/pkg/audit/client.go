package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CommandCheckRequest 命令检查请求
type CommandCheckRequest struct {
	StreamID   string `json:"stream_id"`
	Command    string `json:"command"`
	WorkingDir string `json:"working_dir"`
}

// CommandCheckResponse 命令检查响应
type CommandCheckResponse struct {
	Success bool   `json:"success"`
	Data    struct {
		Blocked bool   `json:"blocked"`
		Reason  string `json:"reason"`
		Action  string `json:"action"` // block/warn/log
	} `json:"data"`
	Error string `json:"error"`
}

// CommandRecordRequest 命令记录请求
type CommandRecordRequest struct {
	StreamID   string `json:"stream_id"`
	Command    string `json:"command"`
	WorkingDir string `json:"working_dir"`
	ExitCode   int    `json:"exit_code"`
}

// CommonResponse 通用响应
type CommonResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// Client 审计客户端
type Client struct {
	dashboardURL string
	httpClient   *http.Client
	token        string
}

// NewClient 创建审计客户端
func NewClient(dashboardURL, token string) *Client {
	return &Client{
		dashboardURL: dashboardURL,
		token:        token,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// CheckCommand 检查命令是否被拦截
func (c *Client) CheckCommand(streamID, command, workingDir string) (blocked bool, reason string, action string, err error) {
	req := CommandCheckRequest{
		StreamID:   streamID,
		Command:    command,
		WorkingDir: workingDir,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return false, "", "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.dashboardURL+"/api/v1/terminal/check-command", bytes.NewReader(data))
	if err != nil {
		return false, "", "", fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		// 网络错误时不阻止命令执行
		return false, "", "", nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", "", fmt.Errorf("read response: %w", err)
	}

	var checkResp CommandCheckResponse
	if err := json.Unmarshal(body, &checkResp); err != nil {
		return false, "", "", fmt.Errorf("unmarshal response: %w", err)
	}

	if !checkResp.Success {
		return false, "", "", fmt.Errorf("api error: %s", checkResp.Error)
	}

	return checkResp.Data.Blocked, checkResp.Data.Reason, checkResp.Data.Action, nil
}

// RecordCommand 记录命令执行（异步）
func (c *Client) RecordCommand(streamID, command, workingDir string, exitCode int) {
	go func() {
		req := CommandRecordRequest{
			StreamID:   streamID,
			Command:    command,
			WorkingDir: workingDir,
			ExitCode:   exitCode,
		}

		data, err := json.Marshal(req)
		if err != nil {
			return
		}

		httpReq, err := http.NewRequest("POST", c.dashboardURL+"/api/v1/terminal/record-command", bytes.NewReader(data))
		if err != nil {
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		if c.token != "" {
			httpReq.Header.Set("Authorization", "Bearer "+c.token)
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			return
		}
		defer resp.Body.Close()
	}()
}
