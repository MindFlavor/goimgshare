package config

import (
	"encoding/json"
	"io"
)

type AuthProvider struct {
	ClientID  string
	Secret    string
	ReturnURL string
}

type Config struct {
	Port                           int
	InternalHTTPFilesPath          string
	CacheInternalHTTPFiles         bool
	LogInternalHTTPFilesAccess     bool
	SharedFoldersConfigurationFile string
	ThumbnailCacheFolder           string
	SmallThumbnailSize             uint
	AverageThumbnailSize           uint
	Google                         *AuthProvider
	Facebook                       *AuthProvider
	Github                         *AuthProvider
}

// Save allows you to save the struct
func (config *Config) Save(w io.Writer) {
	json.NewEncoder(w).Encode(config)
}

// Load loads the config from a stream
func Load(r io.Reader) *Config {
	var config Config
	json.NewDecoder(r).Decode(&config)

	return &config
}
