package mobile

import (
	"os"
	"runtime"
	"strings"
)

// Platform represents the detected platform
type Platform int

const (
	PlatformDesktop Platform = iota
	PlatformAndroid
	PlatformIOS
)

// isAndroidRuntime checks for Android environment markers
// Note: gomobile builds with GOOS=linux even on Android
func isAndroidRuntime() bool {
	// Check for Android environment variable
	if os.Getenv("ANDROID_ROOT") != "" {
		return true
	}
	// Check if we can access Android cache dir
	if dir, err := os.UserCacheDir(); err == nil {
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

// DetectPlatform returns the current platform
func DetectPlatform() Platform {
	gos := runtime.GOOS
	switch gos {
	case "ios":
		return PlatformIOS
	case "android":
		return PlatformAndroid
	default:
		// gomobile uses GOOS=linux for Android builds
		if isAndroidRuntime() {
			return PlatformAndroid
		}
		return PlatformDesktop
	}
}

// IsMobile returns true if running on a mobile platform
func IsMobile() bool {
	return DetectPlatform() != PlatformDesktop
}

// IsAndroid returns true if running on Android
func IsAndroid() bool {
	return DetectPlatform() == PlatformAndroid
}

// IsIOS returns true if running on iOS
func IsIOS() bool {
	return DetectPlatform() == PlatformIOS
}

// GetPlatformName returns a string name for the platform
func GetPlatformName() string {
	switch DetectPlatform() {
	case PlatformAndroid:
		return "android"
	case PlatformIOS:
		return "ios"
	default:
		return "desktop"
	}
}

// GetOS returns the runtime GOOS
func GetOS() string {
	return runtime.GOOS
}

// GetArch returns the runtime architecture
func GetArch() string {
	return runtime.GOARCH
}

// IsTouchDevice returns true if the platform supports touch input
func IsTouchDevice() bool {
	return IsMobile() || strings.Contains(strings.ToLower(GetOS()), "android") || strings.Contains(strings.ToLower(GetOS()), "ios")
}
