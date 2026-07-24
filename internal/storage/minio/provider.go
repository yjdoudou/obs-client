package minio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"obs-client/internal/connection"
	"obs-client/internal/storage"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Provider struct {
	client *minio.Client
}

type ExtraConfig struct {
	UseSSL bool `json:"useSSL"`
}

func init() {
	storage.RegisterProvider("minio", NewProvider)
}

func NewProvider(conn *connection.Connection) (storage.StorageProvider, error) {
	var extraConfig ExtraConfig
	if conn.ExtraConfig != "" {
		if err := json.Unmarshal([]byte(conn.ExtraConfig), &extraConfig); err != nil {
			return nil, fmt.Errorf("解析额外配置失败: %w", err)
		}
	}

	endpoint := conn.Endpoint
	if endpoint == "" {
		endpoint = "localhost:9000"
	}

	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conn.AccessKeyID, conn.SecretAccessKey, ""),
		Secure: extraConfig.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("创建MinIO客户端失败: %w", err)
	}

	return &Provider{client: client}, nil
}

func (p *Provider) TestConnection() (bool, error) {
	_, err := p.client.ListBuckets(context.Background())
	if err != nil {
		return false, err
	}
	return true, nil
}

func (p *Provider) Close() {
}

func (p *Provider) ListBuckets() ([]storage.Bucket, error) {
	result, err := p.client.ListBuckets(context.Background())
	if err != nil {
		return nil, err
	}

	var buckets []storage.Bucket
	for _, b := range result {
		buckets = append(buckets, storage.Bucket{
			Name:         b.Name,
			CreationDate: b.CreationDate.Format(time.RFC3339),
			Location:     "",
		})
	}
	return buckets, nil
}

func (p *Provider) CreateBucket(bucketName string) error {
	return p.client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
}

func (p *Provider) DeleteBucket(bucketName string) error {
	return p.client.RemoveBucket(context.Background(), bucketName)
}

func (p *Provider) ListObjects(bucketName, prefix, delimiter string) (*storage.ListObjectsResult, error) {
	var objects []storage.Object
	var folders []string

	for obj := range p.client.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: delimiter == "",
	}) {
		if obj.Err != nil {
			return nil, obj.Err
		}

		if obj.Key == "" {
			continue
		}

		if strings.HasSuffix(obj.Key, "/") && obj.Size == 0 {
			folders = append(folders, obj.Key)
		} else {
			objects = append(objects, storage.Object{
				Key:          obj.Key,
				Size:         obj.Size,
				LastModified: obj.LastModified.Format(time.RFC3339),
				StorageClass: obj.StorageClass,
				ETag:         strings.Trim(obj.ETag, "\""),
			})
		}
	}

	return &storage.ListObjectsResult{
		Objects: objects,
		Folders: folders,
	}, nil
}

func (p *Provider) UploadObject(bucketName, objectKey, localFilePath string, progressFn storage.ProgressFunc) error {
	cleanPath := filepath.Clean(localFilePath)

	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("文件不存在: %s", cleanPath)
		}
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	if progressFn != nil {
		progressFn(0, fileInfo.Size())
	}

	_, err = p.client.FPutObject(context.Background(), bucketName, objectKey, cleanPath, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("上传失败: %w", err)
	}

	if progressFn != nil {
		progressFn(fileInfo.Size(), fileInfo.Size())
	}

	return nil
}

func (p *Provider) DownloadObject(bucketName, objectKey, localFilePath string, progressFn storage.ProgressFunc) error {
	cleanPath := filepath.Clean(localFilePath)

	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if progressFn != nil {
		stat, err := p.client.StatObject(context.Background(), bucketName, objectKey, minio.StatObjectOptions{})
		if err != nil {
			return fmt.Errorf("获取文件元数据失败: %w", err)
		}
		total := stat.Size

		resp, err := p.client.GetObject(context.Background(), bucketName, objectKey, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}
		defer resp.Close()

		file, err := os.Create(cleanPath)
		if err != nil {
			return fmt.Errorf("创建文件失败: %w", err)
		}
		defer file.Close()

		var transferred int64
		buf := make([]byte, 8192)
		for {
			n, err := resp.Read(buf)
			if n > 0 {
				transferred += int64(n)
				if _, writeErr := file.Write(buf[:n]); writeErr != nil {
					return fmt.Errorf("写入文件失败: %w", writeErr)
				}
				progressFn(transferred, total)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("读取响应失败: %w", err)
			}
		}
	} else {
		err := p.client.FGetObject(context.Background(), bucketName, objectKey, cleanPath, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}
	}

	return nil
}

func (p *Provider) DownloadDirectory(bucketName, prefix, localDir string, progressFn storage.ProgressFunc) error {
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

	result, err := p.ListObjects(bucketName, prefix, "")
	if err != nil {
		return err
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
			return err
		}

		var fileProgressFn func(int64, int64)
		if progressFn != nil {
			startSize := downloadedSize
			fileProgressFn = func(fileTransferred int64, fileTotal int64) {
				progressFn(startSize+fileTransferred, totalSize)
			}
		}

		if err := p.DownloadObject(bucketName, obj.Key, localFilePath, fileProgressFn); err != nil {
			return err
		}

		downloadedSize += obj.Size
	}

	return nil
}

func (p *Provider) DeleteObject(bucketName, objectKey string) error {
	return p.client.RemoveObject(context.Background(), bucketName, objectKey, minio.RemoveObjectOptions{})
}

func (p *Provider) CopyObject(srcBucket, srcKey, dstBucket, dstKey string) error {
	resp, err := p.client.GetObject(context.Background(), srcBucket, srcKey, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("获取源对象失败: %w", err)
	}
	defer resp.Close()

	stat, err := p.client.StatObject(context.Background(), srcBucket, srcKey, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("获取源对象元数据失败: %w", err)
	}

	_, err = p.client.PutObject(context.Background(), dstBucket, dstKey, resp, stat.Size, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("复制对象失败: %w", err)
	}

	return nil
}

func (p *Provider) MoveObject(srcBucket, srcKey, dstBucket, dstKey string) error {
	err := p.CopyObject(srcBucket, srcKey, dstBucket, dstKey)
	if err != nil {
		return err
	}
	return p.DeleteObject(srcBucket, srcKey)
}

func (p *Provider) GetObjectContent(bucketName, objectKey string) ([]byte, string, error) {
	resp, err := p.client.GetObject(context.Background(), bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}
	defer resp.Close()

	content, err := io.ReadAll(resp)
	if err != nil {
		return nil, "", err
	}

	contentType := http.DetectContentType(content)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return content, contentType, nil
}
