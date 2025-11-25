package model

// CommandCheckRequest represents a request to check if a command should be blocked
type CommandCheckRequest struct {
	StreamID   string `json:"stream_id"`
	Command    string `json:"command"`
	WorkingDir string `json:"working_dir"`
}

// CommandCheckResponse represents the response to a command check request
type CommandCheckResponse struct {
	Blocked bool   `json:"blocked"`
	Reason  string `json:"reason,omitempty"`
	Action  string `json:"action,omitempty"` // block/warn/log
}
