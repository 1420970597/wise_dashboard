package audit

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

// Server 本地 API 服务器
type Server struct {
	client   *Client
	listener net.Listener
	server   *http.Server
	wg       sync.WaitGroup
}

// NewServer 创建本地 API 服务器
func NewServer(client *Client) (*Server, error) {
	// 监听随机端口
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	s := &Server{
		client:   client,
		listener: listener,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/check-command", s.handleCheckCommand)
	mux.HandleFunc("/record-command", s.handleRecordCommand)

	s.server = &http.Server{
		Handler: mux,
	}

	return s, nil
}

// Start 启动服务器
func (s *Server) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.server.Serve(s.listener); err != nil && err != http.ErrServerClosed {
			log.Printf("audit server error: %v", err)
		}
	}()
}

// Stop 停止服务器
func (s *Server) Stop() error {
	if err := s.server.Close(); err != nil {
		return err
	}
	s.wg.Wait()
	return nil
}

// GetURL 获取服务器 URL
func (s *Server) GetURL() string {
	return fmt.Sprintf("http://%s", s.listener.Addr().String())
}

// handleCheckCommand 处理命令检查请求
func (s *Server) handleCheckCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CommandCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 调用 Dashboard API
	blocked, reason, action, err := s.client.CheckCommand(req.StreamID, req.Command, req.WorkingDir)
	if err != nil {
		// 出错时不阻止命令执行
		blocked = false
		reason = ""
		action = ""
	}

	// 返回符合 wrapper.sh 期望的格式
	resp := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"blocked": blocked,
			"reason":  reason,
			"action":  action,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleRecordCommand 处理命令记录请求
func (s *Server) handleRecordCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CommandRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 异步记录命令
	s.client.RecordCommand(req.StreamID, req.Command, req.WorkingDir, req.ExitCode)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
