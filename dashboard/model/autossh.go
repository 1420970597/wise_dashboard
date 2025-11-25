package model

import "time"

const (
	AutoSSHMappingTypeLocal  = "local"  // 本地转发 -L
	AutoSSHMappingTypeRemote = "remote" // 远程转发 -R
)

const (
	AutoSSHStatusStopped = "stopped"
	AutoSSHStatusRunning = "running"
	AutoSSHStatusError   = "error"
)

type AutoSSH struct {
	Common
	Name        string    `json:"name"`
	ServerID    uint64    `json:"server_id" gorm:"index:idx_server_mapping_port"`
	MappingType string    `json:"mapping_type" gorm:"index:idx_server_mapping_port"` // local or remote
	SourcePort  int       `json:"source_port" gorm:"index:idx_server_mapping_port"`
	TargetHost  string    `json:"target_host"`
	TargetPort  int       `json:"target_port"`
	Enabled     bool      `json:"enabled"`
	Status      string    `json:"status"`
	LastError   string    `json:"last_error,omitempty"`
	LastStartAt time.Time `json:"last_start_at,omitempty"`
}

type AutoSSHForm struct {
	Name        string `json:"name" validate:"required"`
	ServerID    uint64 `json:"server_id" validate:"required"`
	MappingType string `json:"mapping_type" validate:"required,oneof=local remote"`
	SourcePort  int    `json:"source_port" validate:"required,min=1,max=65535"`
	TargetHost  string `json:"target_host" validate:"required"`
	TargetPort  int    `json:"target_port" validate:"required,min=1,max=65535"`
	Enabled     bool   `json:"enabled"`
}

type TaskAutoSSH struct {
	Action      string            `json:"action"` // start, stop, status
	MappingID   uint64            `json:"mapping_id"`
	MappingType string            `json:"mapping_type"`
	SourcePort  int               `json:"source_port"`
	TargetHost  string            `json:"target_host"`
	TargetPort  int               `json:"target_port"`
	SSHHost     string            `json:"ssh_host"`      // SSH 服务器地址，格式：user@host:port
	SSHOptions  map[string]string `json:"ssh_options,omitempty"`
}

type AutoSSHStatusReport struct {
	MappingID uint64 `json:"mapping_id"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
	PID       int    `json:"pid,omitempty"`
}
