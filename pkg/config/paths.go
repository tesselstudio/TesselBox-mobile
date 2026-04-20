package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// isAndroid returns true if running on Android
// Note: gomobile builds with GOOS=linux, so we check for Android environment markers
func isAndroid() bool {
	// Check for Android environment variable
	if os.Getenv("ANDROID_ROOT") != "" {
		return true
	}
	// Check if we can access Android cache dir (primary method)
	if dir, err := os.UserCacheDir(); err == nil {
		// Android cache dir typically contains "cache" in the path
		if strings.Contains(dir, "/data/data/") || strings.Contains(dir, "/cache/") {
			return true
		}
	}
	// Check for Android-specific paths
	if _, err := os.Stat("/system/build.prop"); err == nil {
		return true
	}
	return runtime.GOOS == "android"
}

// GetTesselboxDir returns the system's tesselbox storage directory
func GetTesselboxDir() string {
	// On Android, use the app's private storage
	if isAndroid() {
		// Try cache dir first, then config dir
		if dir, err := os.UserCacheDir(); err == nil {
			return filepath.Join(dir, "tesselbox")
		}
		if dir, err := os.UserConfigDir(); err == nil {
			return filepath.Join(dir, "tesselbox")
		}
		// Fallback to temp dir
		return filepath.Join(os.TempDir(), "tesselbox")
	}
	// For desktop Linux - use home directory
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".tesselbox")
	}
	// Fallback
	return "/var/lib/tesselbox"
}

// GetSavesDir returns the saves directory
func GetSavesDir() string {
	return filepath.Join(GetTesselboxDir(), "saves")
}

// GetWorldSaveDir returns the save directory for a specific world
func GetWorldSaveDir(worldName string) string {
	return filepath.Join(GetSavesDir(), worldName)
}

// GetWorldsDir returns the worlds directory
func GetWorldsDir() string {
	return filepath.Join(GetTesselboxDir(), "worlds")
}

// GetSkinsDir returns the skins directory
func GetSkinsDir() string {
	return filepath.Join(GetTesselboxDir(), "skins")
}

// GetChestFile returns the path to the chests file for a world
func GetChestFile(worldName string) string {
	return filepath.Join(GetWorldSaveDir(worldName), "chests.json")
}

// EnsureDirectories creates all necessary directories
func EnsureDirectories() error {
	dirs := []string{
		GetTesselboxDir(),
		GetSavesDir(),
		GetWorldsDir(),
		GetSkinsDir(),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
