package huawei

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"obs-client/internal/connection"
	"obs-client/internal/storage"

	obsSDK "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

type Provider struct {
	ObsClient *obsSDK.ObsClient
}

func init() {
	storage.RegisterProvider("huawei", NewProvider)
}

func NewProvider(conn *connection.Connection) (storage.StorageProvider, error) {
	endpoint := conn.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://obs.%s.myhuaweicloud.com", conn.Region)
	}

	obsClient, err := obsSDK.New(
		conn.AccessKeyID,
		conn.SecretAccessKey,
		endpoint,
	)
	if err != nil {
		return nil, err
	}

	return &Provider{ObsClient: obsClient}, nil
}

func (p *Provider) TestConnection() (bool, error) {
	_, err := p.ObsClient.ListBuckets(nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (p *Provider) Close() {
	if p.ObsClient != nil {
		p.ObsClient.Close()
	}
}

func (p *Provider) ListBuckets() ([]storage.Bucket, error) {
	output, err := p.ObsClient.ListBuckets(nil)
	if err != nil {
		return nil, err
	}

	var buckets []storage.Bucket
	for _, b := range output.Buckets {
		buckets = append(buckets, storage.Bucket{
			Name:         b.Name,
			CreationDate: b.CreationDate.Format(time.RFC3339),
			Location:     b.Location,
		})
	}
	return buckets, nil
}

func (p *Provider) CreateBucket(bucketName string) error {
	input := &obsSDK.CreateBucketInput{
		Bucket: bucketName,
	}
	_, err := p.ObsClient.CreateBucket(input)
	return err
}

func (p *Provider) DeleteBucket(bucketName string) error {
	_, err := p.ObsClient.DeleteBucket(bucketName)
	return err
}

func (p *Provider) ListObjects(bucketName, prefix, delimiter string) (*storage.ListObjectsResult, error) {
	input := &obsSDK.ListObjectsInput{}
	input.Bucket = bucketName
	input.Prefix = prefix
	input.Delimiter = delimiter

	output, err := p.ObsClient.ListObjects(input)
	if err != nil {
		return nil, err
	}

	var objects []storage.Object
	for _, c := range output.Contents {
		objects = append(objects, storage.Object{
			Key:          c.Key,
			Size:         c.Size,
			LastModified: c.LastModified.Format(time.RFC3339),
			StorageClass: string(c.StorageClass),
			ETag:         c.ETag,
		})
	}

	var folders []string
	for _, cp := range output.CommonPrefixes {
		folders = append(folders, cp)
	}

	return &storage.ListObjectsResult{
		Objects: objects,
		Folders: folders,
	}, nil
}

type progressListener struct {
	fn func(int64, int64)
}

func (l *progressListener) ProgressChanged(event *obsSDK.ProgressEvent) {
	if l.fn != nil {
		l.fn(event.ConsumedBytes, event.TotalBytes)
	}
}

func (p *Provider) UploadObject(bucketName, objectKey, localFilePath string, progressFn storage.ProgressFunc) error {
	input := &obsSDK.PutFileInput{}
	input.Bucket = bucketName
	input.Key = objectKey

	cleanPath := filepath.Clean(localFilePath)
	input.SourceFile = cleanPath

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

	_, err = p.ObsClient.PutFile(input, obsSDK.WithProgress(&progressListener{fn: progressFn}))
	if err != nil {
		return fmt.Errorf("上传失败: %w", err)
	}
	return nil
}

func (p *Provider) DownloadObject(bucketName, objectKey, localFilePath string, progressFn storage.ProgressFunc) error {
	input := &obsSDK.DownloadFileInput{}
	input.Bucket = bucketName
	input.Key = objectKey

	cleanPath := filepath.Clean(localFilePath)
	input.DownloadFile = cleanPath

	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if progressFn != nil {
		_, err := p.ObsClient.DownloadFile(input, obsSDK.WithProgress(&progressListener{fn: progressFn}))
		if err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}
		return nil
	}

	_, err := p.ObsClient.DownloadFile(input)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
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

	input := &obsSDK.ListObjectsInput{}
	input.Bucket = bucketName
	input.Prefix = prefix

	output, err := p.ObsClient.ListObjects(input)
	if err != nil {
		return err
	}

	var totalSize int64
	for _, content := range output.Contents {
		if !strings.HasSuffix(content.Key, "/") {
			totalSize += content.Size
		}
	}

	var downloadedSize int64

	for _, content := range output.Contents {
		if strings.HasSuffix(content.Key, "/") {
			continue
		}

		relPath := content.Key
		if prefix != "" {
			relPath = strings.TrimPrefix(content.Key, prefix)
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

		if err := p.DownloadObject(bucketName, content.Key, localFilePath, fileProgressFn); err != nil {
			return err
		}

		downloadedSize += content.Size
	}

	return nil
}

func (p *Provider) DeleteObject(bucketName, objectKey string) error {
	input := &obsSDK.DeleteObjectInput{}
	input.Bucket = bucketName
	input.Key = objectKey

	_, err := p.ObsClient.DeleteObject(input)
	return err
}

func (p *Provider) CopyObject(srcBucket, srcKey, dstBucket, dstKey string) error {
	input := &obsSDK.CopyObjectInput{}
	input.Bucket = dstBucket
	input.Key = dstKey
	input.CopySourceBucket = srcBucket
	input.CopySourceKey = srcKey

	_, err := p.ObsClient.CopyObject(input)
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
	input := &obsSDK.GetObjectInput{}
	input.Bucket = bucketName
	input.Key = objectKey

	output, err := p.ObsClient.GetObject(input)
	if err != nil {
		return nil, "", err
	}
	defer output.Body.Close()

	content, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := output.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return content, contentType, nil
}
