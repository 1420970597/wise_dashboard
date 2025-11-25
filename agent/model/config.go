package model

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/go-uuid"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"sigs.k8s.io/yaml"
)

//go:generate go run gen/gen.go -type=AgentConfig
type AgentConfig struct {
	Debug bool `koanf:"debug" json:"debug"`

	Server       string `koanf:"server" json:"server"`               // 服务器地址
	ClientSecret string `koanf:"client_secret" json:"client_secret"` // 客户端密钥
	UUID         string `koanf:"uuid" json:"uuid"`

	HardDrivePartitionAllowlist []string        `koanf:"hard_drive_partition_allowlist" json:"hard_drive_partition_allowlist,omitempty"`
	NICAllowlist                map[string]bool `koanf:"nic_allowlist" json:"nic_allowlist,omitempty"`
	DNS                         []string        `koanf:"dns" json:"dns,omitempty"`
	GPU                         bool            `koanf:"gpu" json:"gpu"`                                         // 是否检查GPU
	Temperature                 bool            `koanf:"temperature" json:"temperature"`                         // 是否检查温度
	SkipConnectionCount         bool            `koanf:"skip_connection_count" json:"skip_connection_count"`     // 跳过连接数检查
	SkipProcsCount              bool            `koanf:"skip_procs_count" json:"skip_procs_count"`               // 跳过进程数量检查
	DisableAutoUpdate           bool            `koanf:"disable_auto_update" json:"disable_auto_update"`         // 关闭自动更新
	DisableForceUpdate          bool            `koanf:"disable_force_update" json:"disable_force_update"`       // 关闭强制更新
	DisableCommandExecute       bool            `koanf:"disable_command_execute" json:"disable_command_execute"` // 关闭命令执行
	ReportDelay                 uint32          `koanf:"report_delay" json:"report_delay"`                       // 报告间隔
	TLS                         bool            `koanf:"tls" json:"tls"`                                         // 是否使用TLS加密传输至服务端
	InsecureTLS                 bool            `koanf:"insecure_tls" json:"insecure_tls"`                       // 是否禁用证书检查
	UseIPv6CountryCode          bool            `koanf:"use_ipv6_country_code" json:"use_ipv6_country_code"`     // 默认优先展示IPv6旗帜
	UseGiteeToUpgrade           bool            `koanf:"use_gitee_to_upgrade" json:"use_gitee_to_upgrade"`       // 强制从Gitee获取更新
	DisableNat                  bool            `koanf:"disable_nat" json:"disable_nat"`                         // 关闭内网穿透
	DisableSendQuery            bool            `koanf:"disable_send_query" json:"disable_send_query"`           // 关闭发送TCP/ICMP/HTTP请求
	IPReportPeriod              uint32          `koanf:"ip_report_period" json:"ip_report_period"`               // IP上报周期
	SelfUpdatePeriod            uint32          `koanf:"self_update_period" json:"self_update_period"`           // 自动更新周期
	CustomIPApi                 []string        `koanf:"custom_ip_api" json:"custom_ip_api,omitempty"`           // 自定义 IP API                      // 重载间隔

	// 审计配置
	AuditEnabled      bool   `koanf:"audit_enabled" json:"audit_enabled"`           // 是否启用终端审计
	AuditDashboardURL string `koanf:"audit_dashboard_url" json:"audit_dashboard_url"` // Dashboard URL for audit
	AuditToken        string `koanf:"audit_token" json:"audit_token"`               // 审计认证 Token

	k        *koanf.Koanf `json:"-"`
	filePath string       `json:"-"`
}

// Read 从给定的文件目录加载配置文件
func (c *AgentConfig) Read(path string) error {
	c.k = koanf.New("")
	c.filePath = path
	saveOnce := sync.OnceValue(c.Save)

	if _, err := os.Stat(path); err == nil {
		err = c.k.Load(file.Provider(path), new(kubeyaml))
		if err != nil {
			return err
		}
	} else {
		defer saveOnce()
	}

	err := c.k.Load(env.Provider("NZ_", "", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "NZ_"))
	}), nil)
	if err != nil {
		return err
	}

	err = c.k.Unmarshal("", c)
	if err != nil {
		return err
	}

	if c.UUID == "" {
		if uuid, err := uuid.GenerateUUID(); err == nil {
			c.UUID = uuid
			defer saveOnce()
		} else {
			return fmt.Errorf("generate UUID failed: %v", err)
		}
	}

	return ValidateConfig(c, false)
}

func (c *AgentConfig) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	dir := filepath.Dir(c.filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	// 检查是否是首次创建配置文件
	_, err = os.Stat(c.filePath)
	isFirstTime := os.IsNotExist(err)

	// 如果是首次创建，添加审计配置的完整示例
	if isFirstTime {
		// 在基础配置后添加审计配置说明
		auditTemplate := `
# ==================== 终端审计配置 ====================
# 终端审计功能会自动启用，Dashboard 地址会从 server 配置自动推断
# 如果需要自定义配置，可以修改以下项：
#
#   audit_enabled: false              # 设为 false 可禁用审计
#   audit_dashboard_url: "http://..."  # 自定义 Dashboard 地址（可选）
#   audit_token: "your-token"          # 认证 Token（建议配置）
#
# 默认行为：自动从 server 地址推断 Dashboard 地址（将端口改为 8008）
# ========================================================
`
		data = append(data, []byte(auditTemplate)...)
	}

	return os.WriteFile(c.filePath, data, 0600)
}

func ValidateConfig(c *AgentConfig, isRemoteEdit bool) error {
	if c.ReportDelay == 0 {
		c.ReportDelay = 3
	}

	if c.IPReportPeriod == 0 {
		c.IPReportPeriod = 1800
	} else if c.IPReportPeriod < 30 {
		c.IPReportPeriod = 30
	}

	if c.ReportDelay < 1 || c.ReportDelay > 4 {
		return errors.New("report-delay ranges from 1-4")
	}

	if !isRemoteEdit {
		if c.Server == "" {
			return errors.New("server address should not be empty")
		}

		if c.ClientSecret == "" {
			return errors.New("client_secret must be specified")
		}

		if _, err := uuid.ParseUUID(c.UUID); err != nil {
			return err
		}
	}

	return nil
}

type kubeyaml struct{}

// Unmarshal parses the given YAML bytes.
func (k *kubeyaml) Unmarshal(b []byte) (map[string]interface{}, error) {
	var out map[string]interface{}
	if err := yaml.Unmarshal(b, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// Marshal marshals the given config map to YAML bytes.
func (k *kubeyaml) Marshal(o map[string]interface{}) ([]byte, error) {
	return yaml.Marshal(o)
}
