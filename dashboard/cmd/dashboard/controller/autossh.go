package controller

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/jinzhu/copier"

	"github.com/nezhahq/nezha/model"
	"github.com/nezhahq/nezha/proto"
	"github.com/nezhahq/nezha/service/singleton"
)

// checkPortConflict 检查端口是否已被占用
func checkPortConflict(serverID uint64, mappingType string, sourcePort int, excludeID uint64) error {
	var count int64
	query := singleton.DB.Model(&model.AutoSSH{}).
		Where("server_id = ? AND mapping_type = ? AND source_port = ?",
			serverID, mappingType, sourcePort)

	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return singleton.Localizer.ErrorT(
			"port %d is already in use for %s mapping on this server",
			sourcePort, mappingType)
	}
	return nil
}

// checkAutoSSHPermission 检查用户是否有权限访问AutoSSH
func checkAutoSSHPermission(c *gin.Context, autoSSHID uint64) bool {
	auth, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return false
	}
	user := auth.(*model.User)

	// 管理员有所有权限
	if user.Role == model.RoleAdmin {
		return true
	}

	// 检查用户是否有权访问关联的服务器
	var autoSSH model.AutoSSH
	if err := singleton.DB.First(&autoSSH, autoSSHID).Error; err != nil {
		return false
	}

	// 使用服务器权限检查
	return checkServerPermission(c, autoSSH.ServerID)
}

// List AutoSSH mappings
// @Summary List AutoSSH mappings
// @Schemes
// @Description List AutoSSH port mappings
// @Security BearerAuth
// @Tags auth required
// @Produce json
// @Success 200 {object} model.CommonResponse[[]model.AutoSSH]
// @Router /autossh [get]
func listAutoSSH(c *gin.Context) ([]*model.AutoSSH, error) {
	auth, ok := c.Get(model.CtxKeyAuthorizedUser)
	if !ok {
		return nil, singleton.Localizer.ErrorT("unauthorized")
	}
	user := auth.(*model.User)

	slist := singleton.AutoSSHShared.GetSortedList()

	var mappings []*model.AutoSSH
	if err := copier.Copy(&mappings, &slist); err != nil {
		return nil, err
	}

	// 如果不是管理员，只返回有权限的服务器上的 AutoSSH
	if user.Role != model.RoleAdmin {
		// 获取用户有权限的服务器ID列表
		var userServers []model.UserServer
		singleton.DB.Where("user_id = ?", user.ID).Find(&userServers)

		serverIDMap := make(map[uint64]bool)
		for _, us := range userServers {
			serverIDMap[us.ServerID] = true
		}

		var filteredList []*model.AutoSSH
		for _, a := range mappings {
			// 检查多对多关联或旧的 UserID 字段
			if serverIDMap[a.ServerID] {
				filteredList = append(filteredList, a)
			} else {
				// 向后兼容：检查服务器的 UserID
				var server model.Server
				if err := singleton.DB.First(&server, a.ServerID).Error; err == nil {
					if server.UserID == user.ID {
						filteredList = append(filteredList, a)
					}
				}
			}
		}
		return filteredList, nil
	}

	return mappings, nil
}

// Add AutoSSH mapping
// @Summary Add AutoSSH mapping
// @Security BearerAuth
// @Schemes
// @Description Add AutoSSH port mapping
// @Tags auth required
// @Accept json
// @param request body model.AutoSSHForm true "AutoSSH Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[uint64]
// @Router /autossh [post]
func createAutoSSH(c *gin.Context) (uint64, error) {
	var af model.AutoSSHForm
	var a model.AutoSSH

	if err := c.ShouldBindJSON(&af); err != nil {
		return 0, err
	}

	if _, ok := singleton.ServerShared.Get(af.ServerID); ok {
		if !checkServerPermission(c, af.ServerID) {
			return 0, singleton.Localizer.ErrorT("permission denied")
		}
	}

	// 检查端口冲突
	if err := checkPortConflict(af.ServerID, af.MappingType, af.SourcePort, 0); err != nil {
		return 0, err
	}

	uid := getUid(c)

	a.UserID = uid
	a.Name = af.Name
	a.ServerID = af.ServerID
	a.MappingType = af.MappingType
	a.SourcePort = af.SourcePort
	a.TargetHost = af.TargetHost
	a.TargetPort = af.TargetPort
	a.Enabled = af.Enabled
	a.Status = model.AutoSSHStatusStopped

	if err := singleton.DB.Create(&a).Error; err != nil {
		return 0, newGormError("%v", err)
	}

	singleton.AutoSSHShared.Update(&a)

	// 如果启用，自动启动
	if a.Enabled {
		go startAutoSSHMapping(&a)
	}

	return a.ID, nil
}

// Edit AutoSSH mapping
// @Summary Edit AutoSSH mapping
// @Security BearerAuth
// @Schemes
// @Description Edit AutoSSH port mapping
// @Tags auth required
// @Accept json
// @param id path uint true "Mapping ID"
// @param request body model.AutoSSHForm true "AutoSSH Request"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /autossh/{id} [patch]
func updateAutoSSH(c *gin.Context) (any, error) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}

	var af model.AutoSSHForm
	if err := c.ShouldBindJSON(&af); err != nil {
		return nil, err
	}

	if _, ok := singleton.ServerShared.Get(af.ServerID); ok {
		if !checkServerPermission(c, af.ServerID) {
			return nil, singleton.Localizer.ErrorT("permission denied")
		}
	}

	var a model.AutoSSH
	if err = singleton.DB.First(&a, id).Error; err != nil {
		return nil, singleton.Localizer.ErrorT("mapping id %d does not exist", id)
	}

	if !checkAutoSSHPermission(c, id) {
		return nil, singleton.Localizer.ErrorT("permission denied")
	}

	// 检查端口冲突（排除当前记录）
	if err := checkPortConflict(af.ServerID, af.MappingType, af.SourcePort, id); err != nil {
		return nil, err
	}

	// 如果配置变更，先停止旧的
	if a.Enabled && (a.MappingType != af.MappingType || a.SourcePort != af.SourcePort ||
		a.TargetHost != af.TargetHost || a.TargetPort != af.TargetPort) {
		stopAutoSSHMapping(&a)
	}

	a.Name = af.Name
	a.ServerID = af.ServerID
	a.MappingType = af.MappingType
	a.SourcePort = af.SourcePort
	a.TargetHost = af.TargetHost
	a.TargetPort = af.TargetPort
	a.Enabled = af.Enabled

	if err := singleton.DB.Save(&a).Error; err != nil {
		return 0, newGormError("%v", err)
	}

	singleton.AutoSSHShared.Update(&a)

	// 如果启用，启动新的
	if a.Enabled {
		go startAutoSSHMapping(&a)
	}

	return nil, nil
}

// Start AutoSSH mapping
// @Summary Start AutoSSH mapping
// @Security BearerAuth
// @Schemes
// @Description Start AutoSSH port mapping
// @Tags auth required
// @param id path uint true "Mapping ID"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /autossh/{id}/start [post]
func startAutoSSHHandler(c *gin.Context) (any, error) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}

	var a model.AutoSSH
	if err = singleton.DB.First(&a, id).Error; err != nil {
		return nil, singleton.Localizer.ErrorT("mapping id %d does not exist", id)
	}

	if !checkAutoSSHPermission(c, id) {
		return nil, singleton.Localizer.ErrorT("permission denied")
	}

	go startAutoSSHMapping(&a)

	return nil, nil
}

// Stop AutoSSH mapping
// @Summary Stop AutoSSH mapping
// @Security BearerAuth
// @Schemes
// @Description Stop AutoSSH port mapping
// @Tags auth required
// @param id path uint true "Mapping ID"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /autossh/{id}/stop [post]
func stopAutoSSHHandler(c *gin.Context) (any, error) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}

	var a model.AutoSSH
	if err = singleton.DB.First(&a, id).Error; err != nil {
		return nil, singleton.Localizer.ErrorT("mapping id %d does not exist", id)
	}

	if !checkAutoSSHPermission(c, id) {
		return nil, singleton.Localizer.ErrorT("permission denied")
	}

	go stopAutoSSHMapping(&a)

	return nil, nil
}

// Batch delete AutoSSH mappings
// @Summary Batch delete AutoSSH mappings
// @Security BearerAuth
// @Schemes
// @Description Batch delete AutoSSH port mappings
// @Tags auth required
// @Accept json
// @param request body []uint64 true "id list"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /batch-delete/autossh [post]
func batchDeleteAutoSSH(c *gin.Context) (any, error) {
	var ids []uint64
	if err := c.ShouldBindJSON(&ids); err != nil {
		return nil, err
	}

	// 检查权限
	for _, id := range ids {
		var a model.AutoSSH
		if err := singleton.DB.First(&a, id).Error; err != nil {
			continue
		}
		if !checkAutoSSHPermission(c, id) {
			return nil, singleton.Localizer.ErrorT("permission denied")
		}
		// 删除前先停止
		if a.Enabled {
			stopAutoSSHMapping(&a)
		}
	}

	if err := singleton.DB.Unscoped().Delete(&model.AutoSSH{}, "id in (?)", ids).Error; err != nil {
		return nil, newGormError("%v", err)
	}

	singleton.AutoSSHShared.Delete(ids)
	return nil, nil
}

// startAutoSSHMapping 启动 AutoSSH 端口映射
func startAutoSSHMapping(a *model.AutoSSH) {
	server, ok := singleton.ServerShared.Get(a.ServerID)
	if !ok || server.TaskStream == nil {
		updateAutoSSHStatus(a.ID, model.AutoSSHStatusError, "server not found or not connected")
		return
	}

	taskData, err := json.Marshal(model.TaskAutoSSH{
		Action:      "start",
		MappingID:   a.ID,
		MappingType: a.MappingType,
		SourcePort:  a.SourcePort,
		TargetHost:  a.TargetHost,
		TargetPort:  a.TargetPort,
		SSHHost:     singleton.Conf.AutoSSHHost,
		SSHOptions: map[string]string{
			"ServerAliveInterval": "30",
			"ServerAliveCountMax": "3",
		},
	})
	if err != nil {
		updateAutoSSHStatus(a.ID, model.AutoSSHStatusError, fmt.Sprintf("marshal task error: %v", err))
		return
	}

	if err := server.TaskStream.Send(&proto.Task{
		Type: model.TaskTypeAutoSSH,
		Data: string(taskData),
	}); err != nil {
		updateAutoSSHStatus(a.ID, model.AutoSSHStatusError, fmt.Sprintf("send task error: %v", err))
		return
	}

	updateAutoSSHStatus(a.ID, model.AutoSSHStatusRunning, "")
}

// stopAutoSSHMapping 停止 AutoSSH 端口映射
func stopAutoSSHMapping(a *model.AutoSSH) {
	server, ok := singleton.ServerShared.Get(a.ServerID)
	if !ok || server.TaskStream == nil {
		updateAutoSSHStatus(a.ID, model.AutoSSHStatusStopped, "")
		return
	}

	taskData, err := json.Marshal(model.TaskAutoSSH{
		Action:    "stop",
		MappingID: a.ID,
	})
	if err != nil {
		return
	}

	server.TaskStream.Send(&proto.Task{
		Type: model.TaskTypeAutoSSH,
		Data: string(taskData),
	})

	updateAutoSSHStatus(a.ID, model.AutoSSHStatusStopped, "")
}

// updateAutoSSHStatus 更新 AutoSSH 状态
func updateAutoSSHStatus(id uint64, status string, errorMsg string) {
	var a model.AutoSSH
	if err := singleton.DB.First(&a, id).Error; err != nil {
		return
	}

	a.Status = status
	a.LastError = errorMsg
	if status == model.AutoSSHStatusRunning {
		a.LastStartAt = time.Now()
	}

	singleton.DB.Save(&a)
	singleton.AutoSSHShared.Update(&a)
}
