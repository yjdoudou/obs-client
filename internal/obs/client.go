package obs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"obs-client/internal/db"

	obsSDK "github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

// Client 封装的OBS客户端
type Client struct {
	ObsClient *obsSDK.ObsClient
}

// NewClient 从数据库连接创建一个OBS客户端实例
func NewClient(connID string, location string) (*Client, error) {
	connStatus := db.GetConnectionStatus(connID)
	if connStatus == "" {
		return nil, fmt.Errorf("connection not found")
	}

	// 获取完整连接配置
	var conn db.Connection
	if err := db.DB.First(&conn, "id = ?", connID).Error; err != nil {
		return nil, err
	}

	endpoint := conn.Endpoint
	if endpoint == "" {
		region := conn.Region
		if location != "" {
			region = location
		}
		if region == "" {
			region = "cn-north-4" // fallback
		}
		endpoint = fmt.Sprintf("https://obs.%s.myhuaweicloud.com", region)
	}

	obsClient, err := obsSDK.New(
		conn.AccessKeyID,
		conn.SecretAccessKey,
		endpoint,
	)
	if err != nil {
		return nil, err
	}

	return &Client{ObsClient: obsClient}, nil
}

// TestConnection 测试给定的连接配置是否有效
func TestConnection(conn *db.Connection) (bool, error) {
	endpoint := conn.Endpoint
	if endpoint == "" {
		region := conn.Region
		if region == "" {
			region = "cn-north-4"
		}
		endpoint = fmt.Sprintf("https://obs.%s.myhuaweicloud.com", region)
	}

	obsClient, err := obsSDK.New(
		conn.AccessKeyID,
		conn.SecretAccessKey,
		endpoint,
	)
	if err != nil {
		return false, err
	}
	defer obsClient.Close()

	// 尝试列出桶以验证权限
	_, err = obsClient.ListBuckets(nil)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ListBuckets 列出所有桶
func (c *Client) ListBuckets() ([]obsSDK.Bucket, error) {
	output, err := c.ObsClient.ListBuckets(nil)
	if err != nil {
		return nil, err
	}
	return output.Buckets, nil
}

// CreateBucket 创建桶
func (c *Client) CreateBucket(bucketName string) error {
	input := &obsSDK.CreateBucketInput{
		Bucket: bucketName,
	}
	_, err := c.ObsClient.CreateBucket(input)
	return err
}

// DeleteBucket 删除桶
func (c *Client) DeleteBucket(bucketName string) error {
	_, err := c.ObsClient.DeleteBucket(bucketName)
	return err
}

// ListObjects 列出对象
func (c *Client) ListObjects(bucketName, prefix, delimiter string) ([]obsSDK.Content, []string, error) {
	input := &obsSDK.ListObjectsInput{}
	input.Bucket = bucketName
	input.Prefix = prefix
	input.Delimiter = delimiter

	output, err := c.ObsClient.ListObjects(input)
	if err != nil {
		return nil, nil, err
	}

	var commonPrefixes []string
	for _, p := range output.CommonPrefixes {
		commonPrefixes = append(commonPrefixes, p)
	}

	return output.Contents, commonPrefixes, nil
}

type progressListener struct {
	fn func(int64, int64)
}

func (l *progressListener) ProgressChanged(event *obsSDK.ProgressEvent) {
	if l.fn != nil {
		l.fn(event.ConsumedBytes, event.TotalBytes)
	}
}

// UploadObject 上传对象
func (c *Client) UploadObject(bucketName, objectKey, localFilePath string, progressFn func(int64, int64)) error {
	input := &obsSDK.PutFileInput{}
	input.Bucket = bucketName
	input.Key = objectKey

	// 清理文件路径，确保使用正确的路径分隔符
	cleanPath := filepath.Clean(localFilePath)
	input.SourceFile = cleanPath

	// 检查文件是否存在并获取文件大小
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("文件不存在: %s", cleanPath)
		}
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 如果有进度回调，先报告总大小
	if progressFn != nil {
		progressFn(0, fileInfo.Size())
	}

	_, err = c.ObsClient.PutFile(input, obsSDK.WithProgress(&progressListener{fn: progressFn}))
	if err != nil {
		return fmt.Errorf("上传失败: %w", err)
	}
	return nil
}

// DownloadObject 下载对象
func (c *Client) DownloadObject(bucketName, objectKey, localFilePath string, progressFn func(int64, int64)) error {
	input := &obsSDK.DownloadFileInput{}
	input.Bucket = bucketName
	input.Key = objectKey

	// 清理文件路径，确保使用正确的路径分隔符
	cleanPath := filepath.Clean(localFilePath)
	input.DownloadFile = cleanPath

	// 确保目标目录存在
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if progressFn != nil {
		_, err := c.ObsClient.DownloadFile(input, obsSDK.WithProgress(&progressListener{fn: progressFn}))
		if err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}
		return nil
	}

	_, err := c.ObsClient.DownloadFile(input)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	return nil
}

// DownloadDirectory 递归下载目录
func (c *Client) DownloadDirectory(bucketName, prefix, localDir string, progressFn func(int64, int64)) error {
	// 提取文件夹名称（prefix 的最后一部分）
	folderName := ""
	trimmedPrefix := strings.TrimSuffix(prefix, "/")
	if lastSlash := strings.LastIndex(trimmedPrefix, "/"); lastSlash != -1 {
		folderName = trimmedPrefix[lastSlash+1:]
	} else {
		folderName = trimmedPrefix
	}

	// 如果文件夹名称不为空，则在本地创建同名子目录
	targetDir := localDir
	if folderName != "" {
		targetDir = filepath.Join(localDir, folderName)
	}

	input := &obsSDK.ListObjectsInput{}
	input.Bucket = bucketName
	input.Prefix = prefix

	output, err := c.ObsClient.ListObjects(input)
	if err != nil {
		return err
	}

	// 计算总大小
	var totalSize int64
	for _, content := range output.Contents {
		if !strings.HasSuffix(content.Key, "/") {
			totalSize += content.Size
		}
	}

	var downloadedSize int64

	for _, content := range output.Contents {
		// 过滤掉当前目录或子目录标记对象
		if strings.HasSuffix(content.Key, "/") {
			continue
		}

		// 计算相对路径（相对于原始 prefix）
		relPath := content.Key
		if prefix != "" {
			relPath = strings.TrimPrefix(content.Key, prefix)
		}

		// 拼接本地路径（放入 targetDir 中）
		localFilePath := filepath.Join(targetDir, relPath)

		// 创建缺失的目录
		dir := filepath.Dir(localFilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		// 为文件夹下载创建包装进度回调
		var fileProgressFn func(int64, int64)
		if progressFn != nil {
			// 保存当前文件开始时的已下载大小
			startSize := downloadedSize
			fileProgressFn = func(fileTransferred int64, fileTotal int64) {
				progressFn(startSize+fileTransferred, totalSize)
			}
		}

		// 下载文件
		if err := c.DownloadObject(bucketName, content.Key, localFilePath, fileProgressFn); err != nil {
			return err
		}

		// 文件下载完成后更新已下载大小
		downloadedSize += content.Size
	}

	return nil
}

// DeleteObject 删除对象
func (c *Client) DeleteObject(bucketName, objectKey string) error {
	input := &obsSDK.DeleteObjectInput{}
	input.Bucket = bucketName
	input.Key = objectKey

	_, err := c.ObsClient.DeleteObject(input)
	return err
}

// CopyObject 复制对象
func (c *Client) CopyObject(srcBucket, srcKey, dstBucket, dstKey string) error {
	input := &obsSDK.CopyObjectInput{}
	input.Bucket = dstBucket
	input.Key = dstKey
	input.CopySourceBucket = srcBucket
	input.CopySourceKey = srcKey

	_, err := c.ObsClient.CopyObject(input)
	return err
}

// MoveObject 移动对象
func (c *Client) MoveObject(srcBucket, srcKey, dstBucket, dstKey string) error {
	err := c.CopyObject(srcBucket, srcKey, dstBucket, dstKey)
	if err != nil {
		return err
	}
	return c.DeleteObject(srcBucket, srcKey)
}

// 虚拟文件夹相关结构

type VirtualFolder struct {
	Name     string           `json:"name"`
	Path     string           `json:"path"`
	Children []*VirtualFolder `json:"children,omitempty"`
	Objects  []obsSDK.Content `json:"objects,omitempty"`
}

// ParseObjectKey 将对象Key解析为虚拟路径
func ParseObjectKey(objectKey string, delimiter string) []string {
	if delimiter == "" {
		delimiter = "/"
	}
	return strings.Split(objectKey, delimiter)
}
