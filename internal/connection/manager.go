package connection

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Storage 定义底层存储接口，解耦数据库
type Storage interface {
	GetConnections() ([]*Connection, error)
	CreateConnection(conn *Connection) error
	UpdateConnection(conn *Connection) error
	DeleteConnection(id string) error
	GetConnectionStatus(id string) string
}

// Manager 处理连接逻辑
type Manager struct {
	storage Storage
}

func NewManager(storage Storage) *Manager {
	return &Manager{storage: storage}
}

func (m *Manager) List() ([]*Connection, error) {
	return m.storage.GetConnections()
}

func (m *Manager) Get(id string) (*Connection, error) {
	conns, err := m.storage.GetConnections()
	if err != nil {
		return nil, err
	}
	for _, c := range conns {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, fmt.Errorf("connection not found")
}

func (m *Manager) Save(conn *Connection) error {
	if conn.ID == "" {
		conn.ID = uuid.New().String()
		conn.CreatedAt = time.Now()
		conn.UpdatedAt = time.Now()
		return m.storage.CreateConnection(conn)
	}
	conn.UpdatedAt = time.Now()
	return m.storage.UpdateConnection(conn)
}

func (m *Manager) Delete(id string) error {
	return m.storage.DeleteConnection(id)
}

func (m *Manager) GetStatus(id string) string {
	return m.storage.GetConnectionStatus(id)
}

func (m *Manager) Duplicate(id string) error {
	conn, err := m.Get(id)
	if err != nil {
		return err
	}

	newConn := *conn
	newConn.ID = uuid.New().String()
	newConn.Name = fmt.Sprintf("%s (Copy)", conn.Name)
	newConn.CreatedAt = time.Now()
	newConn.UpdatedAt = time.Now()

	return m.storage.CreateConnection(&newConn)
}
