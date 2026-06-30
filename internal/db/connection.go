package db

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"obs-client/internal/connection"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

var registeredModels []interface{}

// DBConnector 实现 connection.Storage 接口
type DBConnector struct{}

func RegisterModel(model interface{}) {
	registeredModels = append(registeredModels, model)
}

func (d *DBConnector) GetConnections() ([]*connection.Connection, error) {
	return GetConnections()
}

func (d *DBConnector) CreateConnection(conn *connection.Connection) error {
	return CreateConnection(conn)
}

func (d *DBConnector) UpdateConnection(conn *connection.Connection) error {
	return UpdateConnection(conn)
}

func (d *DBConnector) DeleteConnection(id string) error {
	return DeleteConnection(id)
}

func (d *DBConnector) GetConnectionStatus(id string) string {
	return GetConnectionStatus(id)
}

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

	err = database.AutoMigrate(append([]interface{}{&connection.Connection{}, &TransferTask{}, &OperationHistory{}}, registeredModels...)...)
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
func CreateConnection(conn *connection.Connection) error {
	return DB.Create(conn).Error
}

// UpdateConnection 编辑连接
func UpdateConnection(conn *connection.Connection) error {
	conn.UpdatedAt = time.Now()
	return DB.Save(conn).Error
}

// DeleteConnection 删除连接
func DeleteConnection(id string) error {
	return DB.Delete(&connection.Connection{ID: id}).Error
}

// GetConnections 获取连接列表
func GetConnections() ([]*connection.Connection, error) {
	var connections []*connection.Connection
	err := DB.Find(&connections).Error
	return connections, err
}

// GetConnectionStatus 获取连接状态
func GetConnectionStatus(id string) string {
	var conn connection.Connection
	if err := DB.First(&conn, "id = ?", id).Error; err != nil {
		return "disconnected"
	}
	return conn.Status
}

// ClearAllData 清除所有数据
func ClearAllData() error {
	// 获取所有模型
	models := []interface{}{
		&connection.Connection{},
		&TransferTask{},
		&OperationHistory{},
	}

	// 删除所有表
	for _, model := range models {
		if err := DB.Migrator().DropTable(model); err != nil {
			return err
		}
	}

	// 重新自动迁移以重建表结构
	return DB.AutoMigrate(models...)
}
