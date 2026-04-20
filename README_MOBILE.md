# TesselBox Mobile

A gesture-based mobile port of TesselBox for iOS and Android.

## Features

- **Touch Gestures** - Swipe to move, tap to mine/attack, pinch to zoom
- **Dual Input Support** - Works with both touch and keyboard/mouse (desktop compatible)
- **Mobile Platform Detection** - Automatic platform detection for mobile-specific features

## Controls

| Gesture | Action |
|---------|--------|
| Swipe left/right (left half) | Move left/right |
| Swipe up (left half) | Jump |
| Tap (right half) | Mine / Attack |
| Hold (right half) | Place block |
| Pinch | Zoom in/out |
| Two-finger tap | Open inventory |

## Building

### Prerequisites

1. Go 1.21 or later
2. Android SDK (for Android builds)
3. Xcode (for iOS builds, macOS only)

### Initialize gomobile (one-time setup)

```bash
make mobile-init
```

### Build for Android

```bash
# Build APK (for testing)
make android

# Build Android App Bundle (for Play Store)
make android-bundle
```

### Build for iOS (macOS only)

```bash
# Build .app (for testing)
make ios

# Build .xcarchive (for App Store)
make ios-archive
```

## Project Structure

- `pkg/input/touch.go` - Touch gesture detection
- `pkg/mobile/detect.go` - Platform detection utilities
- `cmd/main.go` - Main game loop with touch integration

## Touch Input System

The touch input system (`pkg/input/touch.go`) provides:

- **Zone-based controls**: Left half for movement, right half for actions
- **Gesture recognition**: Swipe detection with configurable thresholds
- **Multi-touch support**: Pinch zoom with two fingers
- **Hold detection**: Long press for placing blocks

## Compatibility

| Platform | Status | Notes |
|----------|--------|-------|
| Android 5.0+ | ✅ Supported | API level 21+ |
| iOS 12+ | ✅ Supported | iPhone and iPad |
| Desktop | ✅ Supported | Touch + keyboard/mouse |

## Differences from Desktop Version

1. **Input**: Gesture-based instead of keyboard/mouse
2. **UI**: Optimized for touch with larger hit targets
3. **Controls**: Swipe movement instead of WASD
4. **Zoom**: Pinch gesture instead of scroll wheel

## Development

The mobile port maintains full compatibility with the desktop version. All existing features (crafting, combat, survival mode, etc.) work with touch controls.

## License

MIT License - same as the original TesselBox project.
