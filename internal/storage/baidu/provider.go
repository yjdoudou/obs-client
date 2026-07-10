package baidu

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"obs-client/internal/connection"
	"obs-client/internal/storage"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"github.com/baidubce/bce-sdk-go/services/bos/api"
)

type Provider struct {
	client *bos.Client
}

func init() {
	storage.RegisterProvider("baidu", NewProvider)
}

func NewProvider(conn *connection.Connection) (storage.StorageProvider, error) {
	endpoint := conn.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.bcebos.com", conn.Region)
	}

	client, err := bos.NewClient(conn.AccessKeyID, conn.SecretAccessKey, endpoint)
	if err != nil {
		return nil, fmt.Errorf("创建百度BOS客户端失败: %w", err)
	}

	return &Provider{client: client}, nil
}

func (p *Provider) TestConnection() (bool, error) {
	_, err := p.client.ListBuckets()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (p *Provider) Close() {
}

func (p *Provider) ListBuckets() ([]storage.Bucket, error) {
	result, err := p.client.ListBuckets()
	if err != nil {
		return nil, err
	}

	var buckets []storage.Bucket
	for _, b := range result.Buckets {
		buckets = append(buckets, storage.Bucket{
			Name:         b.Name,
			CreationDate: b.CreationDate,
			Location:     b.Location,
		})
	}
	return buckets, nil
}

func (p *Provider) CreateBucket(bucketName string) error {
	_, err := p.client.PutBucket(bucketName)
	return err
}

func (p *Provider) DeleteBucket(bucketName string) error {
	err := p.client.DeleteBucket(bucketName)
	return err
}

func (p *Provider) ListObjects(bucketName, prefix, delimiter string) (*storage.ListObjectsResult, error) {
	var objects []storage.Object
	var folders []string

	args := &api.ListObjectsArgs{
		Prefix:    prefix,
		Delimiter: delimiter,
		MaxKeys:   1000,
	}

	result, err := p.client.ListObjects(bucketName, args)
	if err != nil {
		return nil, err
	}

	for _, c := range result.Contents {
		if !strings.HasSuffix(c.Key, "/") {
			objects = append(objects, storage.Object{
				Key:          c.Key,
				Size:         int64(c.Size),
				LastModified: c.LastModified,
				StorageClass: c.StorageClass,
				ETag:         strings.Trim(c.ETag, "\""),
			})
		}
	}

	for _, cp := range result.CommonPrefixes {
		folders = append(folders, fmt.Sprintf("%v", cp))
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

	file, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	body, _ := bce.NewBodyFromReader(file, fileInfo.Size())

	putArgs := &api.PutObjectArgs{
		ContentLength: fileInfo.Size(),
	}

	_, err = p.client.PutObject(bucketName, objectKey, body, putArgs)
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

	resp, err := p.client.GetObject(bucketName, objectKey, nil)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	file, err := os.Create(cleanPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

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
	err := p.client.DeleteObject(bucketName, objectKey)
	return err
}

func (p *Provider) CopyObject(srcBucket, srcKey, dstBucket, dstKey string) error {
	copyArgs := &api.CopyObjectArgs{}
	_, err := p.client.CopyObject(dstBucket, dstKey, srcBucket, srcKey, copyArgs)
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
	resp, err := p.client.GetObject(bucketName, objectKey, nil)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := resp.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return content, contentType, nil
}
