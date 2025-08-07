# Installation

Multiple installation options are available for gh-action-readme.

## 📦 Binary Releases (Recommended)

Download pre-built binaries for your platform from the [latest release](https://github.com/ivuorinen/gh-action-readme/releases/latest).

### Linux x86_64

```bash
curl -L https://github.com/ivuorinen/gh-action-readme/releases/latest/download/gh-action-readme_Linux_x86_64.tar.gz | tar -xz
sudo mv gh-action-readme /usr/local/bin/
```

### macOS x86_64 (Intel)

```bash
curl -L https://github.com/ivuorinen/gh-action-readme/releases/latest/download/gh-action-readme_Darwin_x86_64.tar.gz | tar -xz
sudo mv gh-action-readme /usr/local/bin/
```

### macOS ARM64 (Apple Silicon)

```bash
curl -L https://github.com/ivuorinen/gh-action-readme/releases/latest/download/gh-action-readme_Darwin_arm64.tar.gz | tar -xz
sudo mv gh-action-readme /usr/local/bin/
```

### Windows x86_64

```powershell
# PowerShell
Invoke-WebRequest -Uri "https://github.com/ivuorinen/gh-action-readme/releases/latest/download/gh-action-readme_Windows_x86_64.zip" -OutFile "gh-action-readme.zip"
Expand-Archive gh-action-readme.zip
```

## 🍺 Package Managers

### Homebrew (macOS/Linux)

```bash
brew install ivuorinen/tap/gh-action-readme
```

### Scoop (Windows)

```powershell
scoop bucket add ivuorinen https://github.com/ivuorinen/scoop-bucket.git
scoop install gh-action-readme
```

### Go Install

```bash
go install github.com/ivuorinen/gh-action-readme@latest
```

## 🐳 Docker

Run directly from Docker without installation:

```bash
# Latest release
docker run --rm -v $(pwd):/workspace ghcr.io/ivuorinen/gh-action-readme:latest gen

# Specific version
docker run --rm -v $(pwd):/workspace ghcr.io/ivuorinen/gh-action-readme:v1.0.0 gen
```

### Docker Compose

```yaml
version: '3.8'
services:
  gh-action-readme:
    image: ghcr.io/ivuorinen/gh-action-readme:latest
    volumes:
      - .:/workspace
    working_dir: /workspace
```

## 🔧 From Source

### Prerequisites

- Go 1.24+
- Git

### Build

```bash
git clone https://github.com/ivuorinen/gh-action-readme.git
cd gh-action-readme
make build
```

### Install System-wide

```bash
sudo cp gh-action-readme /usr/local/bin/
```

## ✅ Verify Installation

```bash
gh-action-readme version
gh-action-readme --help
```

## 🔄 Updates

### Binary/Package Manager

- **Homebrew**: `brew upgrade gh-action-readme`
- **Scoop**: `scoop update gh-action-readme`
- **Manual**: Download new binary and replace existing

### Docker

```bash
docker pull ghcr.io/ivuorinen/gh-action-readme:latest
```

### Go Install

```bash
go install github.com/ivuorinen/gh-action-readme@latest
```

## 🚫 Uninstall

### Binary Installation

```bash
sudo rm /usr/local/bin/gh-action-readme
```

### Package Managers

- **Homebrew**: `brew uninstall gh-action-readme`
- **Scoop**: `scoop uninstall gh-action-readme`

### Configuration Files

```bash
# Remove user configuration (optional)
rm -rf ~/.config/gh-action-readme/
```
