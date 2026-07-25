package storage

type Bucket struct {
	Name         string `json:"name"`
	CreationDate string `json:"creationDate"`
	Location     string `json:"location"`
}

type Object struct {
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	LastModified string `json:"lastModified"`
	StorageClass string `json:"storageClass"`
	ETag         string `json:"eTag"`
}

type ListObjectsResult struct {
	Objects []Object `json:"objects"`
	Folders []string `json:"folders"`
}

type ProgressFunc func(transferred int64, total int64)

type StorageProvider interface {
	TestConnection() (bool, error)
	Close()

	ListBuckets() ([]Bucket, error)
	CreateBucket(bucketName string) error
	DeleteBucket(bucketName string) error

	ListObjects(bucketName, prefix, delimiter string) (*ListObjectsResult, error)
	UploadObject(bucketName, objectKey, localFilePath string, progressFn ProgressFunc) error
	UploadDirectory(bucketName, prefix, localDir string, progressFn ProgressFunc) error
	DownloadObject(bucketName, objectKey, localFilePath string, progressFn ProgressFunc) error
	DownloadDirectory(bucketName, prefix, localDir string, progressFn ProgressFunc) error
	DeleteObject(bucketName, objectKey string) error
	DeleteDirectory(bucketName, prefix string) error
	CopyObject(srcBucket, srcKey, dstBucket, dstKey string) error
	MoveObject(srcBucket, srcKey, dstBucket, dstKey string) error
	GetObjectContent(bucketName, objectKey string) ([]byte, string, error)
}
