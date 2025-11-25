package model

import "time"

// TerminalSession 终端会话记录
type TerminalSession struct {
	Common
	UserID           uint64     `json:"user_id" gorm:"index"`
	Username         string     `json:"username"`
	ServerID         uint64     `json:"server_id" gorm:"index"`
	ServerName       string     `json:"server_name"`
	StreamID         string     `json:"stream_id" gorm:"uniqueIndex"`
	StartedAt        time.Time  `json:"started_at" gorm:"index"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	Duration         int        `json:"duration"` // 持续时间（秒）
	CommandCount     int        `json:"command_count"`
	RecordingPath    string     `json:"recording_path,omitempty"`
	RecordingEnabled bool       `json:"recording_enabled"`
}

// TerminalCommand 终端命令执行记录
type TerminalCommand struct {
	Common
	SessionID   uint64    `json:"session_id" gorm:"index"`
	UserID      uint64    `json:"user_id" gorm:"index"`
	ServerID    uint64    `json:"server_id" gorm:"index"`
	Command     string    `json:"command" gorm:"type:text"`
	WorkingDir  string    `json:"working_dir"`
	ExecutedAt  time.Time `json:"executed_at" gorm:"index"`
	ExitCode    int       `json:"exit_code"`
	Blocked     bool      `json:"blocked" gorm:"index"`
	BlockReason string    `json:"block_reason,omitempty"`
}

// TerminalBlacklist 终端命令黑名单
type TerminalBlacklist struct {
	Common
	Pattern     string    `json:"pattern" gorm:"type:text"`
	Description string    `json:"description"`
	Action      string    `json:"action"` // block/warn/log
	Enabled     bool      `json:"enabled" gorm:"index"`
	CreatedBy   uint64    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// TerminalAuditConfig 终端审计配置
type TerminalAuditConfig struct {
	RecordingEnabled   bool   `json:"recording_enabled"`
	RecordingServers   string `json:"recording_servers" gorm:"type:text"` // JSON array of server IDs
	RetentionDays      int    `json:"retention_days"`                     // 0 = 永久保留
	MaxRecordingSize   int64  `json:"max_recording_size"`                 // MB
	CompressionEnabled bool   `json:"compression_enabled"`
}
