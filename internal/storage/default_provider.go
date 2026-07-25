package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DefaultUploadDirectory(provider StorageProvider, bucketName, prefix, localDir string, progressFn ProgressFunc) error {
	cleanDir := filepath.Clean(localDir)

	fileInfo, err := os.Stat(cleanDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("目录不存在: %s", cleanDir)
		}
		return fmt.Errorf("获取目录信息失败: %w", err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("路径不是目录: %s", cleanDir)
	}

	dirName := filepath.Base(cleanDir)

	var totalSize int64
	err = filepath.WalkDir(cleanDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("遍历目录失败: %w", err)
	}

	var uploadedSize int64

	err = filepath.WalkDir(cleanDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(cleanDir, path)
		if err != nil {
			return fmt.Errorf("计算相对路径失败: %w", err)
		}

		objectKey := prefix + dirName + "/" + strings.ReplaceAll(relPath, "\\", "/")

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}

		var fileProgressFn func(int64, int64)
		if progressFn != nil {
			startSize := uploadedSize
			fileProgressFn = func(fileTransferred int64, fileTotal int64) {
				progressFn(startSize+fileTransferred, totalSize)
			}
		}

		if err := provider.UploadObject(bucketName, objectKey, path, fileProgressFn); err != nil {
			return err
		}

		uploadedSize += fileInfo.Size()
		return nil
	})

	if err != nil {
		return fmt.Errorf("上传目录失败: %w", err)
	}

	return nil
}

func DefaultDownloadDirectory(provider StorageProvider, bucketName, prefix, localDir string, progressFn ProgressFunc) error {
	folderName := ""
	trimmedPrefix := strings.TrimSuffix(prefix, "/")
	if lastSlash := strings.LastIndex(trimmedPrefix, "/"); lastSlash != -1 {
		folderName = trimmedPrefix[lastSlash+1:]
	} else {
		folderName = trimmedPrefix
	}

	targetDir := localDir
	if folderName != "" {
		targetDir = filepath.Join(localDir, folderName)
	}

	result, err := provider.ListObjects(bucketName, prefix, "")
	if err != nil {
		return fmt.Errorf("获取目录对象列表失败: %w", err)
	}

	var totalSize int64
	for _, obj := range result.Objects {
		if !strings.HasSuffix(obj.Key, "/") {
			totalSize += obj.Size
		}
	}

	var downloadedSize int64

	for _, obj := range result.Objects {
		if strings.HasSuffix(obj.Key, "/") {
			continue
		}

		relPath := obj.Key
		if prefix != "" {
			relPath = strings.TrimPrefix(obj.Key, prefix)
		}

		localFilePath := filepath.Join(targetDir, relPath)

		dir := filepath.Dir(localFilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}

		var fileProgressFn func(int64, int64)
		if progressFn != nil {
			startSize := downloadedSize
			fileProgressFn = func(fileTransferred int64, fileTotal int64) {
				progressFn(startSize+fileTransferred, totalSize)
			}
		}

		if err := provider.DownloadObject(bucketName, obj.Key, localFilePath, fileProgressFn); err != nil {
			return err
		}

		downloadedSize += obj.Size
	}

	return nil
}

func DefaultDeleteDirectory(provider StorageProvider, bucketName, prefix string) error {
	result, err := provider.ListObjects(bucketName, prefix, "")
	if err != nil {
		return fmt.Errorf("获取目录对象列表失败: %w", err)
	}

	for _, obj := range result.Objects {
		if err := provider.DeleteObject(bucketName, obj.Key); err != nil {
			return fmt.Errorf("删除对象失败 %s: %w", obj.Key, err)
		}
	}

	for _, folder := range result.Folders {
		if err := provider.DeleteObject(bucketName, folder); err != nil {
			return fmt.Errorf("删除文件夹失败 %s: %w", folder, err)
		}
	}

	return nil
}
