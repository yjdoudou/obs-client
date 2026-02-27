package db

import (
	"time"
)

// TransferTask 传输任务结构体
type TransferTask struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	Type            string    `json:"type" gorm:"not null"` // upload/download
	ConnID          string    `json:"connId" gorm:"not null"`
	BucketName      string    `json:"bucketName" gorm:"not null"`
	ObjectKey       string    `json:"objectKey" gorm:"not null"`
	LocalPath       string    `json:"localPath" gorm:"not null"`
	Size            int64     `json:"size" gorm:"not null"`
	TransferredSize int64     `json:"transferredSize" gorm:"default:0"`
	Status          string    `json:"status" gorm:"default:'pending'"`
	Progress        float64   `json:"progress" gorm:"default:0"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// OperationHistory 操作历史记录
type OperationHistory struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Operation  string    `json:"operation" gorm:"not null"`
	ConnID     string    `json:"connId" gorm:"not null"`
	BucketName string    `json:"bucketName"`
	ObjectKey  string    `json:"objectKey"`
	Details    string    `json:"details"`
	Status     string    `json:"status" gorm:"not null"`
	CreatedAt  time.Time `json:"createdAt"`
}
