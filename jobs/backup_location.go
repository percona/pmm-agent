package jobs

// S3LocationConfig contains required properties for accessing S3 Bucket.
type S3LocationConfig struct {
	Endpoint     string
	AccessKey    string
	SecretKey    string
	BucketName   string
	BucketRegion string
}

// BackupLocationConfig groups all backup locations configs.
type BackupLocationConfig struct {
	S3Config *S3LocationConfig
}
