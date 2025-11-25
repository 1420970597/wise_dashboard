package model

// UserServer 用户-服务器多对多关联表
type UserServer struct {
	UserID   uint64 `gorm:"primaryKey;index" json:"user_id"`
	ServerID uint64 `gorm:"primaryKey;index" json:"server_id"`
}

func (UserServer) TableName() string {
	return "user_servers"
}
