package storage

type StorageConfig struct {
	BucketName      string
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	PublicDomain    string // Optional domain
	UsePublicURL    bool   // Use public URL for accessing files
}
