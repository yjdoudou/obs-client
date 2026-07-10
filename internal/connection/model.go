package connection

import "time"

// Connection 连接配置结构体
type Connection struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	Name            string    `json:"name" gorm:"not null"`
	Provider        string    `json:"provider" gorm:"not null"` // 存储提供商类型：huawei/tencent/alibaba/baidu/minio
	AccessKeyID     string    `json:"accessKeyId" gorm:"not null"`
	SecretAccessKey string    `json:"secretAccessKey" gorm:"not null"`
	Region          string    `json:"region" gorm:"not null"`
	Endpoint        string    `json:"endpoint"`
	Status          string    `json:"status" gorm:"default:'disconnected'"`
	Delimiter       string    `json:"delimiter" gorm:"default:'/'"` // 自定义分隔符
	ExtraConfig     string    `json:"extraConfig"`                  // JSON格式，存储提供商特有配置
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
