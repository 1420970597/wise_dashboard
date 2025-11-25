package model

const (
	_ = iota
	TaskTypeHTTPGet
	TaskTypeICMPPing
	TaskTypeTCPPing
	TaskTypeCommand
	TaskTypeTerminal
	TaskTypeUpgrade
	TaskTypeKeepalive
	TaskTypeTerminalGRPC
	TaskTypeNAT
	TaskTypeReportHostInfoDeprecated
	TaskTypeFM
	TaskTypeReportConfig
	TaskTypeApplyConfig
	TaskTypeAutoSSH
)

type TerminalTask struct {
	StreamID string
}

type TaskNAT struct {
	StreamID string
	Host     string
}

type TaskFM struct {
	StreamID string
}

type TaskAutoSSH struct {
	Action      string            `json:"action"` // start, stop, status
	MappingID   uint64            `json:"mapping_id"`
	MappingType string            `json:"mapping_type"` // local or remote
	SourcePort  int               `json:"source_port"`
	TargetHost  string            `json:"target_host"`
	TargetPort  int               `json:"target_port"`
	SSHHost     string            `json:"ssh_host"`      // SSH 服务器地址，格式：user@host:port
	SSHOptions  map[string]string `json:"ssh_options,omitempty"`
}
