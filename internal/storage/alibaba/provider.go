package alibaba

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"obs-client/internal/connection"
	"obs-client/internal/storage"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Provider struct {
	client *oss.Client
}

type ExtraConfig struct {
	SecurityToken string `json:"securityToken"`
}

func init() {
	storage.RegisterProvider("alibaba", NewProvider)
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
		endpoint = fmt.Sprintf("oss-%s.aliyuncs.com", conn.Region)
	}

	client, err := oss.New(endpoint, conn.AccessKeyID, conn.SecretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("创建OSS客户端失败: %w", err)
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
			CreationDate: b.CreationDate.Format(time.RFC3339),
			Location:     b.Location,
		})
	}
	return buckets, nil
}

func (p *Provider) CreateBucket(bucketName string) error {
	return p.client.CreateBucket(bucketName)
}

func (p *Provider) DeleteBucket(bucketName string) error {
	return p.client.DeleteBucket(bucketName)
}

func (p *Provider) ListObjects(bucketName, prefix, delimiter string) (*storage.ListObjectsResult, error) {
	bucket, err := p.client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	marker := ""
	var objects []storage.Object
	var folders []string

	for {
		lsRes, err := bucket.ListObjects(oss.Prefix(prefix), oss.Delimiter(delimiter), oss.Marker(marker), oss.MaxKeys(1000))
		if err != nil {
			return nil, err
		}

		for _, obj := range lsRes.Objects {
			if !strings.HasSuffix(obj.Key, "/") {
				objects = append(objects, storage.Object{
					Key:          obj.Key,
					Size:         obj.Size,
					LastModified: obj.LastModified.Format(time.RFC3339),
					StorageClass: obj.StorageClass,
					ETag:         strings.Trim(obj.ETag, "\""),
				})
			}
		}

		for _, cp := range lsRes.CommonPrefixes {
			folders = append(folders, cp)
		}

		if !lsRes.IsTruncated {
			break
		}
		marker = lsRes.NextMarker
	}

	return &storage.ListObjectsResult{
		Objects: objects,
		Folders: folders,
	}, nil
}

func (p *Provider) UploadObject(bucketName, objectKey, localFilePath string, progressFn storage.ProgressFunc) error {
	bucket, err := p.client.Bucket(bucketName)
	if err != nil {
		return err
	}

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

	err = bucket.PutObjectFromFile(objectKey, cleanPath)
	if err != nil {
		return fmt.Errorf("上传失败: %w", err)
	}

	if progressFn != nil {
		progressFn(fileInfo.Size(), fileInfo.Size())
	}

	return nil
}

func (p *Provider) UploadDirectory(bucketName, prefix, localDir string, progressFn storage.ProgressFunc) error {
	return storage.DefaultUploadDirectory(p, bucketName, prefix, localDir, progressFn)
}

func (p *Provider) DownloadObject(bucketName, objectKey, localFilePath string, progressFn storage.ProgressFunc) error {
	bucket, err := p.client.Bucket(bucketName)
	if err != nil {
		return err
	}

	cleanPath := filepath.Clean(localFilePath)

	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if progressFn != nil {
		headers, err := bucket.GetObjectDetailedMeta(objectKey)
		if err != nil {
			return fmt.Errorf("获取文件元数据失败: %w", err)
		}
		contentLengthStr := headers.Get("Content-Length")
		total, _ := strconv.ParseInt(contentLengthStr, 10, 64)

		resp, err := bucket.GetObject(objectKey)
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
		err = bucket.GetObjectToFile(objectKey, cleanPath)
		if err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}
	}

	return nil
}

func (p *Provider) DownloadDirectory(bucketName, prefix, localDir string, progressFn storage.ProgressFunc) error {
	return storage.DefaultDownloadDirectory(p, bucketName, prefix, localDir, progressFn)
}

func (p *Provider) DeleteObject(bucketName, objectKey string) error {
	bucket, err := p.client.Bucket(bucketName)
	if err != nil {
		return err
	}
	return bucket.DeleteObject(objectKey)
}

func (p *Provider) DeleteDirectory(bucketName, prefix string) error {
	return storage.DefaultDeleteDirectory(p, bucketName, prefix)
}

func (p *Provider) CopyObject(srcBucket, srcKey, dstBucket, dstKey string) error {
	dstBucketObj, err := p.client.Bucket(dstBucket)
	if err != nil {
		return err
	}

	srcURL := fmt.Sprintf("/%s/%s", srcBucket, srcKey)
	_, err = dstBucketObj.CopyObject(dstKey, srcURL)
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
	bucket, err := p.client.Bucket(bucketName)
	if err != nil {
		return nil, "", err
	}

	resp, err := bucket.GetObject(objectKey)
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
