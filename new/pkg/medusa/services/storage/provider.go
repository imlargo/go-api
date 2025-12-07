package storage

type StorageProvider string

const (
	StorageProviderR2 StorageProvider = "r2"
)

func (sp StorageProvider) IsValid() bool {
	switch sp {
	case StorageProviderR2:
		return true
	default:
		return false
	}
}
