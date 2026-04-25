package assets

import (
	"embed"
)

//go:embed config/*
var ConfigFS embed.FS

// GetConfigFile returns a config file from the embedded filesystem
func GetConfigFile(name string) ([]byte, error) {
	return ConfigFS.ReadFile("config/" + name)
}
