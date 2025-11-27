package singleton

import (
	"log"
	"time"

	"github.com/nezhahq/nezha/model"
)

// CleanupTerminalAuditData cleans up old terminal audit data based on retention policy
func CleanupTerminalAuditData() {
	// 从配置读取保留天数，默认90天
	retentionDays := Conf.TerminalRetentionDays
	if retentionDays == 0 {
		retentionDays = 90 // 默认值
	}

	if retentionDays < 0 {
		// 负数表示永久保留，不清理
		return
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	// Delete old terminal sessions
	result := DB.Where("started_at < ?", cutoffTime).Delete(&model.TerminalSession{})
	if result.Error != nil {
		log.Printf("NEZHA>> Failed to cleanup terminal sessions: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("NEZHA>> Cleaned up %d terminal sessions older than %d days", result.RowsAffected, retentionDays)
	}

	// Delete old terminal commands (orphaned or old)
	result = DB.Where("executed_at < ?", cutoffTime).Delete(&model.TerminalCommand{})
	if result.Error != nil {
		log.Printf("NEZHA>> Failed to cleanup terminal commands: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("NEZHA>> Cleaned up %d terminal commands older than %d days", result.RowsAffected, retentionDays)
	}
}

// StartTerminalAuditCleanupTask starts a background task to periodically clean up old audit data
func StartTerminalAuditCleanupTask() {
	// Run cleanup daily at 3 AM
	go func() {
		for {
			now := time.Now()
			// Calculate next 3 AM
			next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			duration := next.Sub(now)

			log.Printf("NEZHA>> Terminal audit cleanup scheduled for %v (in %v)", next, duration)
			time.Sleep(duration)

			CleanupTerminalAuditData()
		}
	}()
}
