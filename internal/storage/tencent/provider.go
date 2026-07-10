package tencent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"obs-client/internal/connection"
	"obs-client/internal/storage"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type Provider struct {
	client *cos.Client
}

type ExtraConfig struct {
	AppID string `json:"appId"`
}

func init() {
	storage.RegisterProvider("tencent", NewProvider)
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
		endpoint = fmt.Sprintf("https://cos.%s.myqcloud.com", conn.Region)
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("解析Endpoint失败: %w", err)
	}

	baseURL := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  conn.AccessKeyID,
			SecretKey: conn.SecretAccessKey,
		},
	})

	return &Provider{client: client}, nil
}

func (p *Provider) TestConnection() (bool, error) {
	_, _, err := p.client.Service.Get(context.Background())
	if err != nil {
		return false, err
	}
	return true, nil
}

func (p *Provider) Close() {
}

func (p *Provider) ListBuckets() ([]storage.Bucket, error) {
	result, _, err := p.client.Service.Get(context.Background())
	if err != nil {
		return nil, err
	}

	var buckets []storage.Bucket
	for _, b := range result.Buckets {
		buckets = append(buckets, storage.Bucket{
			Name:         b.Name,
			CreationDate: b.CreationDate,
			Location:     "",
		})
	}
	return buckets, nil
}

func (p *Provider) CreateBucket(bucketName string) error {
	_, err := p.client.Bucket.Put(context.Background(), nil)
	return err
}

func (p *Provider) DeleteBucket(bucketName string) error {
	_, err := p.client.Bucket.Delete(context.Background())
	return err
}

func (p *Provider) ListObjects(bucketName, prefix, delimiter string) (*storage.ListObjectsResult, error) {
	var marker string
	var objects []storage.Object
	var folders []string

	for {
		opt := &cos.BucketGetOptions{
			Prefix:    prefix,
			Delimiter: delimiter,
			Marker:    marker,
			MaxKeys:   1000,
		}

		result, _, err := p.client.Bucket.Get(context.Background(), opt)
		if err != nil {
			return nil, err
		}

		for _, c := range result.Contents {
			if !strings.HasSuffix(c.Key, "/") {
				objects = append(objects, storage.Object{
					Key:          c.Key,
					Size:         c.Size,
					LastModified: c.LastModified,
					StorageClass: c.StorageClass,
					ETag:         strings.Trim(c.ETag, "\""),
				})
			}
		}

		for _, cp := range result.CommonPrefixes {
			folders = append(folders, cp)
		}

		if !result.IsTruncated {
			break
		}
		marker = result.NextMarker
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

	file, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	if progressFn != nil {
		progressFn(0, fileInfo.Size())
	}

	_, err = p.client.Object.Put(context.Background(), objectKey, file, nil)
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

	file, err := os.Create(cleanPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	resp, err := p.client.Object.Get(context.Background(), objectKey, nil)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if progressFn != nil {
		total := resp.ContentLength
		var transferred int64
		buf := make([]byte, 8192)
		for {
			n, err := resp.Body.Read(buf)
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
		if _, err := io.Copy(file, resp.Body); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
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
	_, err := p.client.Object.Delete(context.Background(), objectKey)
	return err
}

func (p *Provider) CopyObject(srcBucket, srcKey, dstBucket, dstKey string) error {
	srcURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", srcBucket, p.getRegion(), srcKey)
	_, _, err := p.client.Object.Copy(context.Background(), dstKey, srcURL, nil)
	return err
}

func (p *Provider) MoveObject(srcBucket, srcKey, dstBucket, dstKey string) error {
	err := p.CopyObject(srcBucket, srcKey, dstBucket, dstKey)
	if err != nil {
		return err
	}
	return p.DeleteObject(srcBucket, srcKey)
}

func (p *Provider) GetObjectContent(bucketName, objectKey string) ([]byte, string, error) {
	resp, err := p.client.Object.Get(context.Background(), objectKey, nil)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := http.DetectContentType(content)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return content, contentType, nil
}

func (p *Provider) getRegion() string {
	return ""
}
