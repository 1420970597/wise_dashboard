package audit

import (
	"sync"
)

// Config 审计配置
type Config struct {
	Enabled      bool   // 是否启用审计
	DashboardURL string // Dashboard URL
	Token        string // 认证 Token
}

var (
	globalConfig *Config
	configMutex  sync.RWMutex
)

// SetConfig 设置全局配置
func SetConfig(cfg *Config) {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig = cfg
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	if globalConfig == nil {
		return &Config{
			Enabled: false,
		}
	}
	return globalConfig
}

// IsEnabled 检查审计是否启用
func IsEnabled() bool {
	cfg := GetConfig()
	return cfg.Enabled && cfg.DashboardURL != ""
}
