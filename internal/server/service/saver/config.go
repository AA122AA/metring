package saver

type Config struct {
	StoreInterval   int    `json:"storeInterval" yaml:"storeInterval" env:"STORE_INTERVAL" default:"300"`
	FileStoragePath string `json:"fileStoragePath" yaml:"fileStoragePath" env:"FILE_STORAGE_PATH"`
	Restore         bool   `json:"restore" yaml:"restore" env:"RESTORE" default:"true"`
}
