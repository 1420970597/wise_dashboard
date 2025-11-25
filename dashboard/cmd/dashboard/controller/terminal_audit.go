package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nezhahq/nezha/model"
	"github.com/nezhahq/nezha/service/singleton"
)

// Check command against blacklist
// @Summary Check command against blacklist
// @Description Check if a command should be blocked
// @Tags auth required
// @Accept json
// @Param request body model.CommandCheckRequest true "Command Check Request"
// @Produce json
// @Success 200 {object} model.CommandCheckResponse
// @Router /terminal/check-command [post]
func checkCommand(c *gin.Context) (*model.CommandCheckResponse, error) {
	var req struct {
		StreamID   string `json:"stream_id"`
		Command    string `json:"command"`
		WorkingDir string `json:"working_dir"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}


	// Get session info
	var session model.TerminalSession
	if err := singleton.DB.Where("stream_id = ?", req.StreamID).First(&session).Error; err != nil {
		return &model.CommandCheckResponse{
			Blocked: false,
		}, nil
	}


	// Check against blacklist
	var blacklist []model.TerminalBlacklist
	if err := singleton.DB.Where("enabled = ?", true).Find(&blacklist).Error; err != nil {
		return &model.CommandCheckResponse{
			Blocked: false,
		}, nil
	}


	for _, rule := range blacklist {
		matched, err := regexp.MatchString(rule.Pattern, req.Command)
		if err != nil {
			continue
		}

		if matched {
			// Record command (for warn/log actions)
			if rule.Action == "warn" || rule.Action == "log" {
				cmdRecord := &model.TerminalCommand{
					SessionID:   session.ID,
					UserID:      session.UserID,
					ServerID:    session.ServerID,
					Command:     req.Command,
					WorkingDir:  req.WorkingDir,
					ExecutedAt:  time.Now(),
					Blocked:     false,
					BlockReason: rule.Description,
				}
				singleton.DB.Create(cmdRecord)
			}

			// Handle block action
			if rule.Action == "block" {
				// Record blocked command
				cmdRecord := &model.TerminalCommand{
					SessionID:   session.ID,
					UserID:      session.UserID,
					ServerID:    session.ServerID,
					Command:     req.Command,
					WorkingDir:  req.WorkingDir,
					ExecutedAt:  time.Now(),
					Blocked:     true,
					BlockReason: rule.Description,
				}
				singleton.DB.Create(cmdRecord)

				return &model.CommandCheckResponse{
					Blocked: true,
					Reason:  rule.Description,
					Action:  "block",
				}, nil
			}

			// Handle warn action
			if rule.Action == "warn" {
				return &model.CommandCheckResponse{
					Blocked: false,
					Reason:  rule.Description,
					Action:  "warn",
				}, nil
			}
		}
	}

	return &model.CommandCheckResponse{
		Blocked: false,
	}, nil
}

// Record terminal command
// @Summary Record terminal command
// @Description Record a command executed in terminal
// @Tags auth required
// @Accept json
// @Param request body model.TerminalCommand true "Terminal Command"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /terminal/record-command [post]
func recordCommand(c *gin.Context) (any, error) {
	var req struct {
		StreamID   string `json:"stream_id"`
		Command    string `json:"command"`
		WorkingDir string `json:"working_dir"`
		ExitCode   int    `json:"exit_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}


	// Get session info
	var session model.TerminalSession
	if err := singleton.DB.Where("stream_id = ?", req.StreamID).First(&session).Error; err != nil {
		return nil, err
	}


	// Record command
	cmdRecord := &model.TerminalCommand{
		SessionID:  session.ID,
		UserID:     session.UserID,
		ServerID:   session.ServerID,
		Command:    req.Command,
		WorkingDir: req.WorkingDir,
		ExecutedAt: time.Now(),
		ExitCode:   req.ExitCode,
		Blocked:    false,
	}

	if err := singleton.DB.Create(cmdRecord).Error; err != nil {
		return nil, err
	}


	// Update session command count
	singleton.DB.Model(&session).Update("command_count", session.CommandCount+1)


	return nil, nil
}

// List terminal sessions
// @Summary List terminal sessions
// @Description List terminal sessions with pagination
// @Security BearerAuth
// @Tags auth required
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param user_id query uint64 false "Filter by user ID"
// @Param server_id query uint64 false "Filter by server ID"
// @Produce json
// @Success 200 {object} model.CommonResponse[[]model.TerminalSession]
// @Router /terminal/sessions [get]
func listTerminalSessions(c *gin.Context) (any, error) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID := c.Query("user_id")
	serverID := c.Query("server_id")

	query := singleton.DB.Model(&model.TerminalSession{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if serverID != "" {
		query = query.Where("server_id = ?", serverID)
	}

	var total int64
	query.Count(&total)

	var sessions []model.TerminalSession
	offset := (page - 1) * pageSize
	if err := query.Order("started_at DESC").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, err
	}

	return gin.H{
		"sessions": sessions,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}, nil
}

// List terminal commands
// @Summary List terminal commands
// @Description List terminal commands with pagination
// @Security BearerAuth
// @Tags auth required
// @Param session_id query uint64 false "Filter by session ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Produce json
// @Success 200 {object} model.CommonResponse[[]model.TerminalCommand]
// @Router /terminal/commands [get]
func listTerminalCommands(c *gin.Context) (any, error) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	sessionID := c.Query("session_id")

	query := singleton.DB.Model(&model.TerminalCommand{})

	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}

	var total int64
	query.Count(&total)

	var commands []model.TerminalCommand
	offset := (page - 1) * pageSize
	if err := query.Order("executed_at DESC").Offset(offset).Limit(pageSize).Find(&commands).Error; err != nil {
		return nil, err
	}

	return gin.H{
		"commands": commands,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}, nil
}

// List terminal blacklist rules
// @Summary List terminal blacklist rules
// @Description List terminal command blacklist rules
// @Security BearerAuth
// @Tags auth required
// @Produce json
// @Success 200 {object} model.CommonResponse[[]model.TerminalBlacklist]
// @Router /terminal/blacklist [get]
func listTerminalBlacklist(c *gin.Context) ([]*model.TerminalBlacklist, error) {
	var rules []*model.TerminalBlacklist
	if err := singleton.DB.Order("created_at DESC").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// Create terminal blacklist rule
// @Summary Create terminal blacklist rule
// @Description Create a new terminal command blacklist rule
// @Security BearerAuth
// @Tags auth required
// @Accept json
// @Param request body model.TerminalBlacklist true "Blacklist Rule"
// @Produce json
// @Success 200 {object} model.CommonResponse[uint64]
// @Router /terminal/blacklist [post]
func createTerminalBlacklist(c *gin.Context) (uint64, error) {
	var rule model.TerminalBlacklist
	if err := c.ShouldBindJSON(&rule); err != nil {
		return 0, err
	}

	uid := getUid(c)
	rule.CreatedBy = uid

	if err := singleton.DB.Create(&rule).Error; err != nil {
		return 0, newGormError("%v", err)
	}

	return rule.ID, nil
}

// Update terminal blacklist rule
// @Summary Update terminal blacklist rule
// @Description Update a terminal command blacklist rule
// @Security BearerAuth
// @Tags auth required
// @Accept json
// @Param id path uint true "Rule ID"
// @Param request body model.TerminalBlacklist true "Blacklist Rule"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /terminal/blacklist/{id} [patch]
func updateTerminalBlacklist(c *gin.Context) (any, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}

	var rule model.TerminalBlacklist
	if err := singleton.DB.First(&rule, id).Error; err != nil {
		return nil, singleton.Localizer.ErrorT("rule id %d does not exist", id)
	}

	var updateData model.TerminalBlacklist
	if err := c.ShouldBindJSON(&updateData); err != nil {
		return nil, err
	}

	rule.Pattern = updateData.Pattern
	rule.Description = updateData.Description
	rule.Action = updateData.Action
	rule.Enabled = updateData.Enabled

	if err := singleton.DB.Save(&rule).Error; err != nil {
		return nil, newGormError("%v", err)
	}

	return nil, nil
}

// Delete terminal blacklist rule
// @Summary Delete terminal blacklist rule
// @Description Delete a terminal command blacklist rule
// @Security BearerAuth
// @Tags auth required
// @Param id path uint true "Rule ID"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /terminal/blacklist/{id} [delete]
func deleteTerminalBlacklist(c *gin.Context) (any, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}

	if err := singleton.DB.Delete(&model.TerminalBlacklist{}, id).Error; err != nil {
		return nil, newGormError("%v", err)
	}

	return nil, nil
}

// Upload terminal recording
// @Summary Upload terminal recording
// @Description Upload a terminal session recording
// @Tags auth required
// @Accept json
// @Param request body model.UploadRecordingRequest true "Upload Recording Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /terminal/upload-recording [post]
func uploadRecording(c *gin.Context) (any, error) {
	var req struct {
		StreamID string `json:"stream_id"`
		Data     []byte `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	// Get session info
	var session model.TerminalSession
	if err := singleton.DB.Where("stream_id = ?", req.StreamID).First(&session).Error; err != nil {
		return nil, err
	}

	// Create recordings directory
	recordingsDir := filepath.Join("data", "recordings")
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		return nil, fmt.Errorf("create recordings dir: %w", err)
	}

	// Save recording file
	filename := fmt.Sprintf("%s.cast.gz", req.StreamID)
	filePath := filepath.Join(recordingsDir, filename)

	if err := os.WriteFile(filePath, req.Data, 0644); err != nil {
		return nil, fmt.Errorf("write recording file: %w", err)
	}

	// Update session with recording path
	session.RecordingPath = filePath
	session.RecordingEnabled = true
	if err := singleton.DB.Save(&session).Error; err != nil {
		return nil, err
	}

	return nil, nil
}

// Download terminal recording
// @Summary Download terminal recording
// @Description Download a terminal session recording
// @Security BearerAuth
// @Tags auth required
// @Param session_id path uint true "Session ID"
// @Produce application/gzip
// @Success 200 {file} binary
// @Router /terminal/recording/{session_id} [get]
func downloadRecording(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid session id"})
		return
	}

	// Get session info
	var session model.TerminalSession
	if err := singleton.DB.First(&session, sessionID).Error; err != nil {
		c.JSON(404, gin.H{"error": "session not found"})
		return
	}

	// Check if recording exists
	if session.RecordingPath == "" || !session.RecordingEnabled {
		c.JSON(404, gin.H{"error": "recording not found"})
		return
	}

	// Check if file exists
	if _, err := os.Stat(session.RecordingPath); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": "recording file not found"})
		return
	}

	// Serve file
	c.Header("Content-Type", "application/gzip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.cast.gz", session.StreamID))
	c.File(session.RecordingPath)
}

// Get recording URL
// @Summary Get recording URL
// @Description Get the URL to download a terminal session recording
// @Security BearerAuth
// @Tags auth required
// @Param session_id path uint true "Session ID"
// @Produce json
// @Success 200 {object} model.CommonResponse[string]
// @Router /terminal/recording-url/{session_id} [get]
func getRecordingURL(c *gin.Context) (string, error) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		return "", err
	}

	// Get session info
	var session model.TerminalSession
	if err := singleton.DB.First(&session, sessionID).Error; err != nil {
		return "", err
	}

	// Check if recording exists
	if session.RecordingPath == "" || !session.RecordingEnabled {
		return "", fmt.Errorf("recording not found")
	}

	// Return download URL
	url := fmt.Sprintf("/api/v1/terminal/recording/%d", sessionID)
	return url, nil
}
