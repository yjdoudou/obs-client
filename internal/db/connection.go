package db

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

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

var DB *gorm.DB

func InitDB() error {
	// 获取用户主目录下的应用数据目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	appDir := filepath.Join(homeDir, ".obs-client")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return err
	}

	dbPath := filepath.Join(appDir, "data.sqlite")

	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return err
	}

	err = database.AutoMigrate(&Connection{}, &TransferTask{}, &OperationHistory{})
	if err != nil {
		log.Printf("Failed to auto migrate database: %v", err)
		return err
	}

	DB = database
	log.Println("Database initialized successfully at", dbPath)
	return nil
}

// 核心接口实现

// CreateConnection 创建连接
func CreateConnection(conn *Connection) error {
	conn.CreatedAt = time.Now()
	conn.UpdatedAt = time.Now()
	return DB.Create(conn).Error
}

// UpdateConnection 编辑连接
func UpdateConnection(conn *Connection) error {
	conn.UpdatedAt = time.Now()
	return DB.Save(conn).Error
}

// DeleteConnection 删除连接
func DeleteConnection(id string) error {
	return DB.Delete(&Connection{ID: id}).Error
}

// GetConnections 获取连接列表
func GetConnections() ([]*Connection, error) {
	var connections []*Connection
	err := DB.Find(&connections).Error
	return connections, err
}

// GetConnectionStatus 获取连接状态
func GetConnectionStatus(id string) string {
	var conn Connection
	if err := DB.First(&conn, "id = ?", id).Error; err != nil {
		return "disconnected"
	}
	return conn.Status
}
