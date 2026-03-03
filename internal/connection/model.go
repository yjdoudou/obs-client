package connection

import "time"

// Connection 连接配置结构体
type Connection struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	Name            string    `json:"name" gorm:"not null"`
	AccessKeyID     string    `json:"accessKeyId" gorm:"not null"`
	SecretAccessKey string    `json:"secretAccessKey" gorm:"not null"`
	Region          string    `json:"region" gorm:"not null"`
	Endpoint        string    `json:"endpoint"`
	Status          string    `json:"status" gorm:"default:'disconnected'"`
	Delimiter       string    `json:"delimiter" gorm:"default:'/'"` // 自定义分隔符
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
