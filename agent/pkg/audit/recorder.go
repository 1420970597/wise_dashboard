package audit

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AsciinemaHeader asciinema 格式头部
type AsciinemaHeader struct {
	Version   int               `json:"version"`
	Width     int               `json:"width"`
	Height    int               `json:"height"`
	Timestamp int64             `json:"timestamp"`
	Env       map[string]string `json:"env,omitempty"`
}

// AsciinemaEvent asciinema 格式事件
type AsciinemaEvent struct {
	Time float64 `json:"time"`
	Type string  `json:"type"`
	Data string  `json:"data"`
}

// Recorder 会话录像器
type Recorder struct {
	streamID  string
	startTime time.Time
	file      *os.File
	gzWriter  *gzip.Writer
	mu        sync.Mutex
	closed    bool
}

// NewRecorder 创建录像器
func NewRecorder(streamID string, width, height int) (*Recorder, error) {
	// 创建临时目录
	tmpDir := os.TempDir()
	recordDir := filepath.Join(tmpDir, "nezha-recordings")
	if err := os.MkdirAll(recordDir, 0755); err != nil {
		return nil, fmt.Errorf("create recording dir: %w", err)
	}

	// 创建录像文件
	filename := filepath.Join(recordDir, fmt.Sprintf("%s-%d.cast.gz", streamID, time.Now().Unix()))
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("create recording file: %w", err)
	}

	// 创建 gzip writer
	gzWriter := gzip.NewWriter(file)

	r := &Recorder{
		streamID:  streamID,
		startTime: time.Now(),
		file:      file,
		gzWriter:  gzWriter,
	}

	// 写入 asciinema 头部
	header := AsciinemaHeader{
		Version:   2,
		Width:     width,
		Height:    height,
		Timestamp: r.startTime.Unix(),
		Env: map[string]string{
			"TERM": "xterm",
		},
	}

	headerData, err := json.Marshal(header)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("marshal header: %w", err)
	}

	if _, err := gzWriter.Write(append(headerData, '\n')); err != nil {
		file.Close()
		return nil, fmt.Errorf("write header: %w", err)
	}

	return r, nil
}

// WriteOutput 记录输出数据
func (r *Recorder) WriteOutput(data []byte) error {
	return r.writeEvent("o", data)
}

// WriteInput 记录输入数据
func (r *Recorder) WriteInput(data []byte) error {
	return r.writeEvent("i", data)
}

// writeEvent 写入事件
func (r *Recorder) writeEvent(eventType string, data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("recorder is closed")
	}

	// 计算相对时间（秒）
	elapsed := time.Since(r.startTime).Seconds()

	// 创建事件
	event := []interface{}{
		elapsed,
		eventType,
		string(data),
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if _, err := r.gzWriter.Write(append(eventData, '\n')); err != nil {
		return fmt.Errorf("write event: %w", err)
	}

	return nil
}

// Close 关闭录像器并返回录像文件路径
func (r *Recorder) Close() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return "", fmt.Errorf("recorder already closed")
	}

	r.closed = true

	// 关闭 gzip writer
	if err := r.gzWriter.Close(); err != nil {
		return "", fmt.Errorf("close gzip writer: %w", err)
	}

	// 获取文件路径
	filePath := r.file.Name()

	// 关闭文件
	if err := r.file.Close(); err != nil {
		return "", fmt.Errorf("close file: %w", err)
	}

	return filePath, nil
}

// GetFilePath 获取录像文件路径
func (r *Recorder) GetFilePath() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.file.Name()
}

// UploadRecording 上传录像到 Dashboard
func (c *Client) UploadRecording(streamID, filePath string) error {
	// 读取录像文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read recording file: %w", err)
	}

	// 创建请求
	url := fmt.Sprintf("%s/api/v1/terminal/upload-recording", c.dashboardURL)

	// 构建请求体
	reqBody := map[string]interface{}{
		"stream_id": streamID,
		"data":      data,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(reqData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// 删除本地文件
	os.Remove(filePath)

	return nil
}
