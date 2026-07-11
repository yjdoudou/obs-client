package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"obs-client/internal/connection"
	"obs-client/internal/db"
	"obs-client/internal/storage"
	_ "obs-client/internal/storage/alibaba"
	_ "obs-client/internal/storage/baidu"
	_ "obs-client/internal/storage/huawei"
	_ "obs-client/internal/storage/minio"
	_ "obs-client/internal/storage/tencent"
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

type Bucket struct {
	Name         string `json:"Name"`
	CreationDate string `json:"CreationDate"`
	Location     string `json:"Location"`
}

type OBSObject struct {
	Key          string `json:"Key"`
	Size         int64  `json:"Size"`
	LastModified string `json:"LastModified"`
	StorageClass string `json:"StorageClass"`
}

type ListObjectsResponse struct {
	Objects []OBSObject `json:"objects"`
	Folders []string    `json:"folders"`
}

type PreviewFileResponse struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
	FileName    string `json:"fileName"`
	Size        int64  `json:"size"`
}

type App struct {
	ctx          context.Context
	connManager  *connection.Manager
	themeManager *theme.Manager
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	if err := db.InitDB(); err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
	}
	a.connManager = connection.NewManager(&db.DBConnector{})
	a.themeManager = theme.NewManager(db.DB)
}

func (a *App) SelectFile() ([]string, error) {
	result, err := wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择文件上传",
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *App) SelectSaveFile(defaultName string) (string, error) {
	return wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "下载文件为...",
		DefaultFilename: defaultName,
	})
}

func (a *App) SelectDirectory() (string, error) {
	return wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择下载到的本地目录",
	})
}

func (a *App) GetConnections() ([]*connection.Connection, error) {
	return a.connManager.List()
}

func (a *App) CreateConnection(conn *connection.Connection) (*connection.Connection, error) {
	if err := a.connManager.Save(conn); err != nil {
		return nil, err
	}
	return conn, nil
}

func (a *App) UpdateConnection(conn *connection.Connection) (bool, error) {
	if err := a.connManager.Save(conn); err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) DeleteConnection(id string) (bool, error) {
	if err := a.connManager.Delete(id); err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) DuplicateConnection(id string) (bool, error) {
	if err := a.connManager.Duplicate(id); err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) TestConnection(conn *connection.Connection) (bool, error) {
	provider, err := storage.NewProvider(conn)
	if err != nil {
		return false, err
	}
	defer provider.Close()
	return provider.TestConnection()
}

func (a *App) getProvider(connID string) (storage.StorageProvider, error) {
	conn, err := a.connManager.Get(connID)
	if err != nil {
		return nil, err
	}
	return storage.NewProvider(conn)
}

func (a *App) ListBuckets(connID string) ([]Bucket, error) {
	provider, err := a.getProvider(connID)
	if err != nil {
		return nil, err
	}
	defer provider.Close()

	buckets, err := provider.ListBuckets()
	if err != nil {
		return nil, err
	}

	result := make([]Bucket, len(buckets))
	for i, b := range buckets {
		result[i] = Bucket{
			Name:         b.Name,
			CreationDate: b.CreationDate,
			Location:     b.Location,
		}
	}
	return result, nil
}

func (a *App) CreateBucket(connID string, location string, bucketName string) (bool, error) {
	provider, err := a.getProvider(connID)
	if err != nil {
		return false, err
	}
	defer provider.Close()

	err = provider.CreateBucket(bucketName)
	return err == nil, err
}

func (a *App) DeleteBucket(connID string, location string, bucketName string) (bool, error) {
	provider, err := a.getProvider(connID)
	if err != nil {
		return false, err
	}
	defer provider.Close()

	err = provider.DeleteBucket(bucketName)
	return err == nil, err
}

func (a *App) ListObjects(connID string, location string, bucketName string, prefix string, delimiter string) (*ListObjectsResponse, error) {
	provider, err := a.getProvider(connID)
	if err != nil {
		return nil, err
	}
	defer provider.Close()

	result, err := provider.ListObjects(bucketName, prefix, delimiter)
	if err != nil {
		return nil, err
	}

	objects := make([]OBSObject, len(result.Objects))
	for i, c := range result.Objects {
		objects[i] = OBSObject{
			Key:          c.Key,
			Size:         c.Size,
			LastModified: c.LastModified,
			StorageClass: c.StorageClass,
		}
	}

	return &ListObjectsResponse{
		Objects: objects,
		Folders: result.Folders,
	}, nil
}

func (a *App) UploadFile(connID string, location string, bucketName string, objectKey string, localFilePath string, taskID string) (string, error) {
	provider, err := a.getProvider(connID)
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "transfer-error", map[string]interface{}{
			"id":    taskID,
			"error": err.Error(),
		})
		return "", err
	}
	defer provider.Close()

	progressFn := func(transferred int64, total int64) {
		wailsRuntime.EventsEmit(a.ctx, "transfer-progress", map[string]interface{}{
			"id":          taskID,
			"transferred": transferred,
			"total":       total,
		})
	}

	err = provider.UploadObject(bucketName, objectKey, localFilePath, progressFn)
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

func (a *App) DownloadFile(connID string, location string, bucketName string, objectKey string, localFilePath string, taskID string) (string, error) {
	provider, err := a.getProvider(connID)
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "transfer-error", map[string]interface{}{
			"id":    taskID,
			"error": err.Error(),
		})
		return "", err
	}
	defer provider.Close()

	progressFn := func(transferred int64, total int64) {
		wailsRuntime.EventsEmit(a.ctx, "transfer-progress", map[string]interface{}{
			"id":          taskID,
			"transferred": transferred,
			"total":       total,
		})
	}

	err = provider.DownloadObject(bucketName, objectKey, localFilePath, progressFn)
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

func (a *App) DownloadDirectory(connID string, location string, bucketName string, prefix string, localDir string, taskID string) error {
	provider, err := a.getProvider(connID)
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "transfer-error", map[string]interface{}{
			"id":    taskID,
			"error": err.Error(),
		})
		return err
	}
	defer provider.Close()

	progressFn := func(transferred int64, total int64) {
		wailsRuntime.EventsEmit(a.ctx, "transfer-progress", map[string]interface{}{
			"id":          taskID,
			"transferred": transferred,
			"total":       total,
		})
	}

	err = provider.DownloadDirectory(bucketName, prefix, localDir, progressFn)
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

func (a *App) DeleteObject(connID string, location string, bucketName string, objectKey string) (bool, error) {
	provider, err := a.getProvider(connID)
	if err != nil {
		return false, err
	}
	defer provider.Close()

	err = provider.DeleteObject(bucketName, objectKey)
	return err == nil, err
}

func (a *App) PreviewFile(connID string, location string, bucketName string, objectKey string) (*PreviewFileResponse, error) {
	const maxPreviewSize = 10 * 1024 * 1024

	provider, err := a.getProvider(connID)
	if err != nil {
		return nil, err
	}
	defer provider.Close()

	content, contentType, err := provider.GetObjectContent(bucketName, objectKey)
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

func (a *App) CopyObject(connID string, location string, srcBucket string, srcKey string, dstBucket string, dstKey string) (bool, error) {
	provider, err := a.getProvider(connID)
	if err != nil {
		return false, err
	}
	defer provider.Close()

	err = provider.CopyObject(srcBucket, srcKey, dstBucket, dstKey)
	return err == nil, err
}

func (a *App) MoveObject(connID string, location string, srcBucket string, srcKey string, dstBucket string, dstKey string) (bool, error) {
	provider, err := a.getProvider(connID)
	if err != nil {
		return false, err
	}
	defer provider.Close()

	err = provider.MoveObject(srcBucket, srcKey, dstBucket, dstKey)
	return err == nil, err
}

func (a *App) ClearAppCache() (bool, error) {
	err := db.ClearAllData()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) GetAllThemes() []theme.ThemeConfig {
	return a.themeManager.GetAllThemes()
}

func (a *App) GetCurrentTheme(userID string) (string, error) {
	return a.themeManager.GetCurrentThemeJSON(userID)
}

func (a *App) SaveTheme(userID string, themeID string, isDark bool) (bool, error) {
	if err := a.themeManager.SaveUserTheme(userID, themeID, isDark); err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) SaveBackgroundImage(userID string, base64Image string) (bool, error) {
	if err := a.themeManager.SaveBackgroundImage(userID, base64Image); err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) GetBackgroundInfo(userID string) (theme.BackgroundInfo, error) {
	return a.themeManager.GetBackgroundInfo(userID)
}

func (a *App) SaveOverlayOpacity(userID string, opacity float64) (bool, error) {
	if err := a.themeManager.SaveOverlayOpacity(userID, opacity); err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) ClearBackgroundImage(userID string) (bool, error) {
	if err := a.themeManager.ClearBackgroundImage(userID); err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) EditFile(connID string, location string, bucketName string, objectKey string) (string, error) {
	conn, err := a.connManager.Get(connID)
	if err != nil {
		return "", err
	}
	provider, err := storage.NewProvider(conn)
	if err != nil {
		return "", err
	}
	defer provider.Close()

	fileName := objectKey
	if idx := strings.LastIndex(objectKey, "/"); idx != -1 {
		fileName = objectKey[idx+1:]
	}

	tempDir, err := os.MkdirTemp("", "obs-edit-*")
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败: %v", err)
	}

	localFilePath := filepath.Join(tempDir, fileName)

	err = provider.DownloadObject(bucketName, objectKey, localFilePath, nil)
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

				editProvider, err := storage.NewProvider(conn)
				if err != nil {
					fmt.Printf("创建Provider失败: %v\n", err)
					continue
				}

				err = editProvider.UploadObject(bucketName, objectKey, localFilePath, nil)
				editProvider.Close()

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

	return fmt.Sprintf("文件已在系统应用中打开，编辑后将自动同步到云存储\n临时文件: %s", localFilePath), nil
}

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
