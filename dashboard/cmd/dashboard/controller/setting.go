package controller

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nezhahq/nezha/model"
	"github.com/nezhahq/nezha/service/singleton"
)

// List settings
// @Summary List settings
// @Schemes
// @Description List settings
// @Security BearerAuth
// @Tags common
// @Produce json
// @Success 200 {object} model.CommonResponse[model.SettingResponse]
// @Router /setting [get]
func listConfig(c *gin.Context) (*model.SettingResponse, error) {
	u, authorized := c.Get(model.CtxKeyAuthorizedUser)
	var isAdmin bool
	if authorized {
		user := u.(*model.User)
		isAdmin = user.Role.IsAdmin()
	}

	config := *singleton.Conf
	config.Language = strings.ReplaceAll(config.Language, "_", "-")

	conf := model.SettingResponse{
		Config: model.Setting{
			ConfigForGuests:                config.ConfigForGuests,
			ConfigDashboard:                config.ConfigDashboard,
			IgnoredIPNotificationServerIDs: config.IgnoredIPNotificationServerIDs,
			Oauth2Providers:                config.Oauth2Providers,
		},
		Version:           singleton.Version,
		FrontendTemplates: singleton.FrontendTemplates,
	}

	if !authorized || !isAdmin {
		configForGuests := config.ConfigForGuests
		var configDashboard model.ConfigDashboard
		if authorized {
			configDashboard.AgentTLS = singleton.Conf.AgentTLS
			configDashboard.InstallHost = singleton.Conf.InstallHost

			// 为已登录用户生成安装命令
			if singleton.Conf.InstallHost != "" {
				user := u.(*model.User)
				conf.InstallCommand = generateInstallCommand(singleton.Conf.InstallHost, singleton.Conf.AgentTLS, user.AgentSecret)
			}
		}
		conf = model.SettingResponse{
			Config: model.Setting{
				ConfigForGuests: configForGuests,
				ConfigDashboard: configDashboard,
				Oauth2Providers: config.Oauth2Providers,
			},
			Version:           singleton.Version,
			FrontendTemplates: singleton.FrontendTemplates,
			InstallCommand:    conf.InstallCommand,
		}
	} else {
		// 为管理员生成安装命令
		log.Printf("NEZHA>> Admin user detected, InstallHost: %s", singleton.Conf.InstallHost)
		if singleton.Conf.InstallHost != "" {
			user := u.(*model.User)
			log.Printf("NEZHA>> User ID: %d, AgentSecret: %s", user.ID, user.AgentSecret)
			conf.InstallCommand = generateInstallCommand(singleton.Conf.InstallHost, singleton.Conf.AgentTLS, user.AgentSecret)
			log.Printf("NEZHA>> Generated InstallCommand: %s", conf.InstallCommand)
		} else {
			log.Printf("NEZHA>> InstallHost is empty!")
		}
	}

	return &conf, nil
}

// Edit config
// @Summary Edit config
// @Security BearerAuth
// @Schemes
// @Description Edit config
// @Tags admin required
// @Accept json
// @Param body body model.SettingForm true "SettingForm"
// @Produce json
// @Success 200 {object} model.CommonResponse[any]
// @Router /setting [patch]
func updateConfig(c *gin.Context) (any, error) {
	var sf model.SettingForm
	if err := c.ShouldBindJSON(&sf); err != nil {
		return nil, err
	}
	var userTemplateValid bool
	for _, v := range singleton.FrontendTemplates {
		if !userTemplateValid && v.Path == sf.UserTemplate && !v.IsAdmin {
			userTemplateValid = true
		}
		if userTemplateValid {
			break
		}
	}
	if !userTemplateValid {
		return nil, errors.New("invalid user template")
	}

	singleton.Conf.Language = strings.ReplaceAll(sf.Language, "-", "_")

	singleton.Conf.EnableIPChangeNotification = sf.EnableIPChangeNotification
	singleton.Conf.EnablePlainIPInNotification = sf.EnablePlainIPInNotification
	singleton.Conf.Cover = sf.Cover
	singleton.Conf.InstallHost = sf.InstallHost
	singleton.Conf.IgnoredIPNotification = sf.IgnoredIPNotification
	singleton.Conf.IPChangeNotificationGroupID = sf.IPChangeNotificationGroupID
	singleton.Conf.SiteName = sf.SiteName
	singleton.Conf.DNSServers = sf.DNSServers
	singleton.Conf.AutoSSHHost = sf.AutoSSHHost
	singleton.Conf.CustomCode = sf.CustomCode
	singleton.Conf.CustomCodeDashboard = sf.CustomCodeDashboard
	singleton.Conf.WebRealIPHeader = sf.WebRealIPHeader
	singleton.Conf.AgentRealIPHeader = sf.AgentRealIPHeader
	singleton.Conf.AgentTLS = sf.AgentTLS
	singleton.Conf.UserTemplate = sf.UserTemplate

	if err := singleton.Conf.Save(); err != nil {
		return nil, newGormError("%v", err)
	}

	singleton.OnUpdateLang(singleton.Conf.Language)
	return nil, nil
}

// generateInstallCommand 生成 agent 安装命令
func generateInstallCommand(installHost string, agentTLS bool, userSecret string) string {
	// 构建下载 URL
	protocol := "http"
	if agentTLS {
		protocol = "https"
	}
	installScriptURL := fmt.Sprintf("%s://%s/install.sh", protocol, installHost)

	// TLS 参数
	tlsValue := "false"
	if agentTLS {
		tlsValue = "true"
	}

	// 构建完整的安装命令
	installCmd := fmt.Sprintf(
		"curl -L %s -o nezha.sh && chmod +x nezha.sh && sudo NZ_SERVER=%s NZ_CLIENT_SECRET=%s NZ_TLS=%s ./nezha.sh",
		installScriptURL,
		installHost,
		userSecret,
		tlsValue,
	)

	return installCmd
}
