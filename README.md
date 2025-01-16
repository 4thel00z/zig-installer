# ```zig-installer```
![GPL-3 License](./LICENSE)

## What this project is about

A cli for installing the Zig compiler.
It automatically downloads, verifies, and installs Zig from official releases with support for custom installation paths and multiple versions.

## Installation

```bash
go install github.com/4thel00z/zig-installer/...@latest
```


## Usage Examples

### Basic Installation
```bash
sudo zig-installer
```
This will grab the latest master version and install it to `/usr/local`.

### Install Specific Version
```bash
sudo zig-installer --version=0.11.0
```

### Custom Installation Path
```bash
sudo zig-installer \
  --bin-dir=/opt/zig/bin \
  --lib-dir=/opt/zig/lib
```

### Using Environment Variables
```bash
export ZIG_VERSION=0.11.0
export ZIG_BIN_DIR=/opt/zig/bin
export ZIG_LIB_DIR=/opt/zig/lib
sudo zig-installer
```

## Configuration Options

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `--version` | `ZIG_VERSION` | master | Version to install |
| `--bin-dir` | `ZIG_BIN_DIR` | /usr/local/bin | Binary installation path |
| `--lib-dir` | `ZIG_LIB_DIR` | /usr/local/lib | Library installation path |
| `--tar-dest` | `ZIG_TAR_DEST` | /tmp/zig.tar.xz | Download location |
| `--dest` | `ZIG_DEST` | /tmp/zig | Temporary extraction path |
| `--index-url` | `ZIG_INDEX_URL` | ziglang.org/... | Download index URL |

## Features

- 🚀 Fast downloads
- 📦 Proper library installation
- 🔧 Highly configurable
- 🖥️ Cross-platform support
- 🔐 Checksum verification

## Sample Output

```
💡 info: found existing file, checking checksum...
✅ success: existing file matches checksum, skipping download
👉 step: extracting...
👉 step: installing...
👉 step: cleaning up...
✅ success: Zig 0.11.0 installed successfully! 🎉
```
## Troubleshooting

### Permission Errors
```bash
# Run with sudo for system directories
sudo zig-installer
```

### Custom Location Without Root
```bash
# Install to user-owned directory
zig-installer --bin-dir=$HOME/.local/bin --lib-dir=$HOME/.local/lib
```
## License

This project is licensed under the GPL-3 license.
