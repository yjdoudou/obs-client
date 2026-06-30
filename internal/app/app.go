package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"obs-client/internal/connection"
	"obs-client/internal/db"
	"obs-client/internal/obs"
	"obs-client/internal/theme"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Bucket 桶结构镜像
type Bucket struct {
	Name         string `json:"Name"`
	CreationDate string `json:"CreationDate"`
	Location     string `json:"Location"`
}

// OBSObject 对象结构镜像
type OBSObject struct {
	Key          string `json:"Key"`
	Size         int64  `json:"Size"`
	LastModified string `json:"LastModified"`
	StorageClass string `json:"StorageClass"`
}

// ListObjectsResponse 包含对象和公共前缀（文件夹）
type ListObjectsResponse struct {
	Objects []OBSObject `json:"objects"`
	Folders []string    `json:"folders"`
}

// PreviewFileResponse 文件预览响应
type PreviewFileResponse struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
	FileName    string `json:"fileName"`
	Size        int64  `json:"size"`
}

// App struct
type App struct {
	ctx          context.Context
	connManager  *connection.Manager
	themeManager *theme.Manager
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
	// 初始化连接管理器
	a.connManager = connection.NewManager(&db.DBConnector{})
	// 初始化主题管理器
	a.themeManager = theme.NewManager(db.DB)
}

// === Native File Dialogs ===

// SelectFile 选择要上传的本地文件
func (a *App) SelectFile() ([]string, error) {
	result, err := wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择文件上传",
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SelectSaveFile 选择要保存的本地路径
func (a *App) SelectSaveFile(defaultName string) (string, error) {
	return wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "下载文件为...",
		DefaultFilename: defaultName,
	})
}

// SelectDirectory 选择要下载到的本地目录
func (a *App) SelectDirectory() (string, error) {
	return wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择下载到的本地目录",
	})
}

// === 连接管理方法 ===

// GetConnections 获取连接列表
func (a *App) GetConnections() ([]*connection.Connection, error) {
	return a.connManager.List()
}

// CreateConnection 创建连接
func (a *App) CreateConnection(conn *connection.Connection) (*connection.Connection, error) {
	if err := a.connManager.Save(conn); err != nil {
		return nil, err
	}
	return conn, nil
}

// UpdateConnection 更新连接
func (a *App) UpdateConnection(conn *connection.Connection) (bool, error) {
	if err := a.connManager.Save(conn); err != nil {
		return false, err
	}
	return true, nil
}

// DeleteConnection 删除连接
func (a *App) DeleteConnection(id string) (bool, error) {
	if err := a.connManager.Delete(id); err != nil {
		return false, err
	}
	return true, nil
}

// DuplicateConnection 复制连接
func (a *App) DuplicateConnection(id string) (bool, error) {
	if err := a.connManager.Duplicate(id); err != nil {
		return false, err
	}
	return true, nil
}

// TestConnection 测试连接
func (a *App) TestConnection(conn *connection.Connection) (bool, error) {
	return obs.TestConnection(conn)
}

// === OBS 操作方法 ===

// ListBuckets 获取桶列表
func (a *App) ListBuckets(connID string) ([]Bucket, error) {
	conn, err := a.connManager.Get(connID)
	if err != nil {
		return nil, err
	}
	client, err := obs.NewClientFromConfig(conn, "")
	if err != nil {
		return nil, err
	}
	obsBuckets, err := client.ListBuckets()
	if err != nil {
		return nil, err
	}

	result := make([]Bucket, len(obsBuckets))
	for i, b := range obsBuckets {
		result[i] = Bucket{
			Name:         b.Name,
			CreationDate: b.CreationDate.String(),
			Location:     b.Location,
		}
	}
	return result, nil
}

// CreateBucket 创建桶
func (a *App) CreateBucket(connID string, location string, bucketName string) (bool, error) {
	conn, err := a.connManager.Get(connID)
	if err != nil {
		return false, err
	}
	client, err := obs.NewClientFromConfig(conn, location)
	if err != nil {
		return false, err
	}
	err = client.CreateBucket(bucketName)
	return err == nil, err
}

// DeleteBucket 删除桶
func (a *App) DeleteBucket(connID string, location string, bucketName string) (bool, error) {
	conn, err := a.connManager.Get(connID)
	if err != nil {
		return false, err
	}
	client, err := obs.NewClientFromConfig(conn, location)
	if err != nil {
		return false, err
	}
	err = client.DeleteBucket(bucketName)
	return err == nil, err
}

// ListObjects 获取对象列表
func (a *App) ListObjects(connID string, location string, bucketName string, prefix string, delimiter string) (*ListObjectsResponse, error) {
	conn, err := a.connManager.Get(connID)
	if err != nil {
		return nil, err
	}
	client, err := obs.NewClientFromConfig(conn, location)
	if err != nil {
		return nil, err
	}
	contents, commonPrefixes, err := client.ListObjects(bucketName, prefix, delimiter)
	if err != nil {
		return nil, err
	}

	objects := make([]OBSObject, len(contents))
	for i, c := range contents {
		objects[i] = OBSObject{
			Key:          c.Key,
			Size:         c.Size,
			LastModified: c.LastModified.String(),
			StorageClass: string(c.StorageClass),
		}
	}

	return &ListObjectsResponse{
		Objects: objects,
		Folders: commonPrefixes,
	}, nil
}

// UploadFile 上传文件
func (a *App) UploadFile(connID string, location string, bucketName string, objectKey string, localFilePath string, taskID string) (string, error) {
	conn, err := a.connManager.Get(connID)
	if err != nil {
		return "", err
	}
	client, err := obs.NewClientFromConfig(conn, location)
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "transfer-error", map[string]interface{}{
			"id":    taskID,
			"error": err.Error(),
		})
		return "", err
	}

	progressFn := func(transferred int64, total int64) {
		wailsRuntime.EventsEmit(a.ctx, "transfer-progress", map[string]interface{}{
			"id":          taskID,
			"transferred": transferred,
			"total":       total,
		})
	}

	err = client.UploadObject(bucketName, objectKey, localFilePath, progressFn)
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "transfer-error", map[string]interface{}{
			"id":    taskID,
			"error": err.Error(),
		})
		return "", err
	}

	wailsRuntime.EventsEmit(a.ctx, "transfer-complete", map[string]interface{}{
		"id": taskID,
	})
	return objectKey, nil
}

// DownloadFile 下载文件
func (a *App) DownloadFile(connID string, location string, bucketName string, objectKey string, localFilePath string, taskID string) (string, error) {
	conn, err := a.connManager.Get(connID)
	if err != nil {
		return "", err
	}
	client, err := obs.NewClientFromConfig(conn, location)
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "transfer-error", map[string]interface{}{
			"id":    taskID,
			"error": err.Error(),
		})
		return "", err
	}

	progressFn := func(transferred int64, total int64) {
		wailsRuntime.EventsEmit(a.ctx, "transfer-progress", map[string]interface{}{
			"id":          taskID,
			"transferred": transferred,
			"total":       total,
		})
	}

	err = client.DownloadObject(bucketName, objectKey, localFilePath, progressFn)
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "transfer-error", map[string]interface{}{
			"id":    taskID,
			"error": err.Error(),
		})
		return "", err
	}

	wailsRuntime.EventsEmit(a.ctx, "transfer-complete", map[string]interface{}{
		"id": taskID,
	})
	return objectKey, nil
}

// DownloadDirectory 递归下载整个目录
func (a *App) DownloadDirectory(connID string, location string, bucketName string, prefix string, localDir string, taskID string) error {
	conn, err := a.connManager.Get(connID)
	if err != nil {
		return err
	}
	client, err := obs.NewClientFromConfig(conn, location)
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "transfer-error", map[string]interface{}{
			"id":    taskID,
			"error": err.Error(),
		})
		return err
	}

	progressFn := func(transferred int64, total int64) {
		wailsRuntime.EventsEmit(a.ctx, "transfer-progress", map[string]interface{}{
			"id":          taskID,
			"transferred": transferred,
			"total":       total,
		})
	}

	err = client.DownloadDirectory(bucketName, prefix, localDir, progressFn)
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "transfer-error", map[string]interface{}{
			"id":    taskID,
			"error": err.Error(),
		})
		return err
	}

	wailsRuntime.EventsEmit(a.ctx, "transfer-complete", map[string]interface{}{
		"id": taskID,
	})
	return nil
}

// DeleteObject 删除对象
func (a *App) DeleteObject(connID string, location string, bucketName string, objectKey string) (bool, error) {
	conn, err := a.connManager.Get(connID)
	if err != nil {
		return false, err
	}
	client, err := obs.NewClientFromConfig(conn, location)
	if err != nil {
		return false, err
	}
	err = client.DeleteObject(bucketName, objectKey)
	return err == nil, err
}

// PreviewFile 预览文件内容（限制10MB）
func (a *App) PreviewFile(connID string, location string, bucketName string, objectKey string) (*PreviewFileResponse, error) {
	const maxPreviewSize = 10 * 1024 * 1024

	conn, err := a.connManager.Get(connID)
	if err != nil {
		return nil, err
	}
	client, err := obs.NewClientFromConfig(conn, location)
	if err != nil {
		return nil, err
	}

	content, contentType, err := client.GetObjectContent(bucketName, objectKey)
	if err != nil {
		return nil, err
	}

	if len(content) > maxPreviewSize {
		return nil, fmt.Errorf("文件大小 %d 超过预览限制 %d", len(content), maxPreviewSize)
	}

	base64Content := base64.StdEncoding.EncodeToString(content)

	fileName := objectKey
	if idx := strings.LastIndex(objectKey, "/"); idx != -1 {
		fileName = objectKey[idx+1:]
	}

	return &PreviewFileResponse{
		Content:     base64Content,
		ContentType: contentType,
		FileName:    fileName,
		Size:        int64(len(content)),
	}, nil
}

// CopyObject 复制对象
func (a *App) CopyObject(connID string, location string, srcBucket string, srcKey string, dstBucket string, dstKey string) (bool, error) {
	conn, err := a.connManager.Get(connID)
	if err != nil {
		return false, err
	}
	client, err := obs.NewClientFromConfig(conn, location)
	if err != nil {
		return false, err
	}
	err = client.CopyObject(srcBucket, srcKey, dstBucket, dstKey)
	return err == nil, err
}

// MoveObject 移动对象
func (a *App) MoveObject(connID string, location string, srcBucket string, srcKey string, dstBucket string, dstKey string) (bool, error) {
	conn, err := a.connManager.Get(connID)
	if err != nil {
		return false, err
	}
	client, err := obs.NewClientFromConfig(conn, location)
	if err != nil {
		return false, err
	}
	err = client.MoveObject(srcBucket, srcKey, dstBucket, dstKey)
	return err == nil, err
}

// ClearAppCache 清除应用缓存（数据重置）
func (a *App) ClearAppCache() (bool, error) {
	err := db.ClearAllData()
	if err != nil {
		return false, err
	}
	return true, nil
}

// === 主题管理方法 ===

// GetAllThemes 获取所有可用主题列表
func (a *App) GetAllThemes() []theme.ThemeConfig {
	return a.themeManager.GetAllThemes()
}

// GetCurrentTheme 获取当前用户主题
func (a *App) GetCurrentTheme(userID string) (string, error) {
	return a.themeManager.GetCurrentThemeJSON(userID)
}

// SaveTheme 设置用户主题
func (a *App) SaveTheme(userID string, themeID string, isDark bool) (bool, error) {
	if err := a.themeManager.SaveUserTheme(userID, themeID, isDark); err != nil {
		return false, err
	}
	return true, nil
}

// SaveBackgroundImage 保存用户背景图片
func (a *App) SaveBackgroundImage(userID string, base64Image string) (bool, error) {
	if err := a.themeManager.SaveBackgroundImage(userID, base64Image); err != nil {
		return false, err
	}
	return true, nil
}

// GetBackgroundInfo 获取用户背景信息
func (a *App) GetBackgroundInfo(userID string) (theme.BackgroundInfo, error) {
	return a.themeManager.GetBackgroundInfo(userID)
}

// SaveOverlayOpacity 保存遮罩透明度
func (a *App) SaveOverlayOpacity(userID string, opacity float64) (bool, error) {
	if err := a.themeManager.SaveOverlayOpacity(userID, opacity); err != nil {
		return false, err
	}
	return true, nil
}

// ClearBackgroundImage 清除用户背景图片
func (a *App) ClearBackgroundImage(userID string) (bool, error) {
	if err := a.themeManager.ClearBackgroundImage(userID); err != nil {
		return false, err
	}
	return true, nil
}

// EditFile 编辑文件：下载到临时目录，用系统应用打开，监听变化并自动上传
func (a *App) EditFile(connID string, location string, bucketName string, objectKey string) (string, error) {
	conn, err := a.connManager.Get(connID)
	if err != nil {
		return "", err
	}
	client, err := obs.NewClientFromConfig(conn, location)
	if err != nil {
		return "", err
	}

	fileName := objectKey
	if idx := strings.LastIndex(objectKey, "/"); idx != -1 {
		fileName = objectKey[idx+1:]
	}

	tempDir, err := os.MkdirTemp("", "obs-edit-*")
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败: %v", err)
	}

	localFilePath := filepath.Join(tempDir, fileName)

	err = client.DownloadObject(bucketName, objectKey, localFilePath, nil)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("下载文件失败: %v", err)
	}

	go func() {
		time.Sleep(5 * time.Second)

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			fmt.Printf("创建文件监听器失败: %v\n", err)
			return
		}
		defer watcher.Close()

		err = watcher.Add(tempDir)
		if err != nil {
			fmt.Printf("添加监听失败: %v\n", err)
			return
		}

		timer := time.NewTimer(0)
		timer.Stop()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if (event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Remove == fsnotify.Remove ||
					event.Op&fsnotify.Rename == fsnotify.Rename) &&
					strings.HasSuffix(event.Name, fileName) {

					timer.Stop()
					timer = time.NewTimer(2 * time.Second)
				}

			case <-timer.C:
				timer.Stop()

				info, err := os.Stat(localFilePath)
				if err != nil {
					continue
				}

				if info.Size() == 0 {
					continue
				}

				err = client.UploadObject(bucketName, objectKey, localFilePath, nil)
				if err != nil {
					fmt.Printf("上传文件失败: %v\n", err)
				} else {
					fmt.Printf("文件已更新: %s\n", objectKey)
					wailsRuntime.EventsEmit(a.ctx, "file-updated", map[string]interface{}{
						"bucketName": bucketName,
						"objectKey":  objectKey,
						"fileName":   fileName,
					})
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("监听错误: %v\n", err)
			}
		}
	}()

	err = openWithSystemApp(localFilePath)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("打开文件失败: %v", err)
	}

	return fmt.Sprintf("文件已在系统应用中打开，编辑后将自动同步到 OBS\n临时文件: %s", localFilePath), nil
}

// openWithSystemApp 使用系统默认应用打开文件
func openWithSystemApp(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer.exe", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return cmd.Start()
}
