package app

import (
	"context"
	"fmt"
	"obs-client/internal/db"
	"obs-client/internal/obs"

	obsSDK "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	// 初始化数据库
	if err := db.InitDB(); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
	}
}

// === Native File Dialogs ===

// SelectFile 选择要上传的本地文件
func (a *App) SelectFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择文件上传",
	})
}

// SelectSaveFile 选择要保存的本地路径
func (a *App) SelectSaveFile(defaultName string) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "下载文件为...",
		DefaultFilename: defaultName,
	})
}

// SelectDirectory 选择要下载到的本地目录
func (a *App) SelectDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择下载到的本地目录",
	})
}

// === 连接管理方法 ===

// GetConnections 获取连接列表
func (a *App) GetConnections() ([]*db.Connection, error) {
	return db.GetConnections()
}

// CreateConnection 创建连接
func (a *App) CreateConnection(conn *db.Connection) (*db.Connection, error) {
	if err := db.CreateConnection(conn); err != nil {
		return nil, err
	}
	return conn, nil
}

// UpdateConnection 更新连接
func (a *App) UpdateConnection(conn *db.Connection) (bool, error) {
	if err := db.UpdateConnection(conn); err != nil {
		return false, err
	}
	return true, nil
}

// DeleteConnection 删除连接
func (a *App) DeleteConnection(id string) (bool, error) {
	if err := db.DeleteConnection(id); err != nil {
		return false, err
	}
	return true, nil
}

// TestConnection 测试连接
func (a *App) TestConnection(conn *db.Connection) (bool, error) {
	return obs.TestConnection(conn)
}

// === OBS 操作方法 ===

// ListBuckets 获取桶列表
func (a *App) ListBuckets(connID string) ([]obsSDK.Bucket, error) {
	client, err := obs.NewClient(connID, "")
	if err != nil {
		return nil, err
	}
	return client.ListBuckets()
}

// CreateBucket 创建桶
func (a *App) CreateBucket(connID string, location string, bucketName string) (bool, error) {
	client, err := obs.NewClient(connID, location)
	if err != nil {
		return false, err
	}
	err = client.CreateBucket(bucketName)
	return err == nil, err
}

// DeleteBucket 删除桶
func (a *App) DeleteBucket(connID string, location string, bucketName string) (bool, error) {
	client, err := obs.NewClient(connID, location)
	if err != nil {
		return false, err
	}
	err = client.DeleteBucket(bucketName)
	return err == nil, err
}

// ListObjectsResponse 包含对象和公共前缀（文件夹）
type ListObjectsResponse struct {
	Objects []obsSDK.Content `json:"objects"`
	Folders []string         `json:"folders"`
}

// ListObjects 获取对象列表
func (a *App) ListObjects(connID string, location string, bucketName string, prefix string, delimiter string) (*ListObjectsResponse, error) {
	client, err := obs.NewClient(connID, location)
	if err != nil {
		return nil, err
	}
	contents, commonPrefixes, err := client.ListObjects(bucketName, prefix, delimiter)
	if err != nil {
		return nil, err
	}
	return &ListObjectsResponse{
		Objects: contents,
		Folders: commonPrefixes,
	}, nil
}

// UploadFile 上传文件
func (a *App) UploadFile(connID string, location string, bucketName string, objectKey string, localFilePath string, taskID string) (string, error) {
	client, err := obs.NewClient(connID, location)
	if err != nil {
		return "", err
	}

	progressFn := func(transferred int64, total int64) {
		runtime.EventsEmit(a.ctx, "transfer-progress", map[string]interface{}{
			"id":          taskID,
			"transferred": transferred,
			"total":       total,
		})
	}

	err = client.UploadObject(bucketName, objectKey, localFilePath, progressFn)
	if err != nil {
		return "", err
	}
	return objectKey, nil
}

// DownloadFile 下载文件
func (a *App) DownloadFile(connID string, location string, bucketName string, objectKey string, localFilePath string, taskID string) (string, error) {
	client, err := obs.NewClient(connID, location)
	if err != nil {
		return "", err
	}

	progressFn := func(transferred int64, total int64) {
		runtime.EventsEmit(a.ctx, "transfer-progress", map[string]interface{}{
			"id":          taskID,
			"transferred": transferred,
			"total":       total,
		})
	}

	err = client.DownloadObject(bucketName, objectKey, localFilePath, progressFn)
	if err != nil {
		return "", err
	}
	return objectKey, nil
}

// DownloadDirectory 递归下载整个目录
func (a *App) DownloadDirectory(connID string, location string, bucketName string, prefix string, localDir string, taskID string) error {
	client, err := obs.NewClient(connID, location)
	if err != nil {
		return err
	}
	return client.DownloadDirectory(bucketName, prefix, localDir)
}

// DeleteObject 删除对象
func (a *App) DeleteObject(connID string, location string, bucketName string, objectKey string) (bool, error) {
	client, err := obs.NewClient(connID, location)
	if err != nil {
		return false, err
	}
	err = client.DeleteObject(bucketName, objectKey)
	return err == nil, err
}

// CopyObject 复制对象
func (a *App) CopyObject(connID string, location string, srcBucket string, srcKey string, dstBucket string, dstKey string) (bool, error) {
	client, err := obs.NewClient(connID, location)
	if err != nil {
		return false, err
	}
	err = client.CopyObject(srcBucket, srcKey, dstBucket, dstKey)
	return err == nil, err
}

// MoveObject 移动对象
func (a *App) MoveObject(connID string, location string, srcBucket string, srcKey string, dstBucket string, dstKey string) (bool, error) {
	client, err := obs.NewClient(connID, location)
	if err != nil {
		return false, err
	}
	err = client.MoveObject(srcBucket, srcKey, dstBucket, dstKey)
	return err == nil, err
}
