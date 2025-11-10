# Building from Source

This guide covers building Valhalla from source for development work.

## Prerequisites

- **Go 1.25 or later** - [Download Go](https://golang.org/dl/)
- **Git** - For cloning the repository
- **Data.nx file** - See [Installation Guide](Installation.md) for conversion
- **MySQL 5.7+** or **MariaDB** - For database

## Quick Start

### Step 1: Install Go

#### Windows

1. Download installer from [golang.org/dl/](https://golang.org/dl/)
2. Run installer
3. Verify installation:
   ```cmd
   go version
   ```

#### Linux

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# Or download latest from golang.org
wget https://go.dev/dl/go1.25.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify
go version
```

#### macOS

```bash
# Using Homebrew
brew install go

# Or download from golang.org

# Verify
go version
```

### Step 2: Clone Repository

```bash
git clone https://github.com/Hucaru/Valhalla.git
cd Valhalla
```

### Step 3: Download Dependencies

```bash
go mod download
```

This downloads all required Go modules specified in `go.mod`.

### Step 4: Build

```bash
go build -v .
```

This creates an executable:
- **Windows**: `Valhalla.exe`
- **Linux/macOS**: `Valhalla`

The `-v` flag shows verbose output to see which packages are being compiled.

### Step 5: Run

Follow the [Local Setup Guide](Local.md) to configure and run the built executable.

## Development Workflow

### Building During Development

For faster iteration during development:

```bash
# Build and run immediately
go run . -type login -config config_login.toml

# Build with race detector (slower but catches concurrency bugs)
go build -race -v .
```

### Project Structure

```
Valhalla/
├── main.go              # Entry point
├── server_login.go      # Login server implementation
├── server_world.go      # World server implementation
├── server_channel.go    # Channel server implementation
├── server_cashshop.go   # Cash shop server implementation
├── server_config.go     # Configuration loading
├── common/              # Shared utilities
├── channel/             # Channel server logic
├── login/               # Login server logic
├── world/               # World server logic
├── cashshop/            # Cash shop logic
├── mnet/                # Network layer
├── mpacket/             # Packet handling
├── nx/                  # NX file reader
├── constant/            # Game constants
├── internal/            # Internal packages
├── scripts/             # NPC scripts (JavaScript)
├── drops.json           # Drop data
├── reactors.json        # Reactor data
└── reactor_drops.json   # Reactor drop data
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./channel/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Formatting

Format code using Go's standard formatter:

```bash
# Format all files
go fmt ./...

# Check formatting without applying changes
go fmt -n ./...
```

### Linting

Use `golangci-lint` for comprehensive linting:

```bash
# Install golangci-lint
# See https://golangci-lint.run/usage/install/

# Run linter
golangci-lint run

# Run with auto-fix
golangci-lint run --fix
```

### Checking for Issues

```bash
# Run go vet (included in go test)
go vet ./...

# Check for common mistakes
go vet -composites=false ./...
```

## Building for Different Platforms

Go supports cross-compilation for different operating systems and architectures.

### Build for Linux (from any OS)

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o Valhalla-linux-amd64 .

# Linux ARM64 (for ARM servers/Raspberry Pi)
GOOS=linux GOARCH=arm64 go build -o Valhalla-linux-arm64 .
```

### Build for Windows (from any OS)

```bash
GOOS=windows GOARCH=amd64 go build -o Valhalla-windows-amd64.exe .
```

### Build for macOS (from any OS)

```bash
# Intel Macs
GOOS=darwin GOARCH=amd64 go build -o Valhalla-darwin-amd64 .

# Apple Silicon (M1/M2)
GOOS=darwin GOARCH=arm64 go build -o Valhalla-darwin-arm64 .
```

### Build Script

Create a build script for all platforms:

```bash
#!/bin/bash
# build-all.sh

platforms=(
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "darwin/amd64"
    "darwin/arm64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a parts <<< "$platform"
    GOOS="${parts[0]}"
    GOARCH="${parts[1]}"
    
    output="Valhalla-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output="${output}.exe"
    fi
    
    echo "Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -o "build/$output" .
done

echo "Build complete! Binaries in build/"
```

Run it:
```bash
chmod +x build-all.sh
./build-all.sh
```

## Build Options

### Optimized Builds

```bash
# Reduce binary size by stripping debug info
go build -ldflags="-s -w" .

# Build with static linking (useful for Docker alpine images)
CGO_ENABLED=0 go build -ldflags="-s -w" .
```

### Debug Builds

```bash
# Build with race detector
go build -race .

# Build with optimizations disabled (easier debugging)
go build -gcflags="all=-N -l" .
```

### Version Information

Embed version information at build time:

```bash
# In main.go, add variables:
# var (
#     Version   = "dev"
#     BuildTime = "unknown"
#     GitCommit = "unknown"
# )

# Build with version info
go build -ldflags="-X main.Version=1.0.0 -X main.BuildTime=$(date -u +%Y%m%d%H%M%S) -X main.GitCommit=$(git rev-parse HEAD)" .
```

## Docker Build

Build the Docker image locally:

```bash
# Build image
docker build -t valhalla:dev .

# Build with specific platform
docker build --platform linux/amd64 -t valhalla:dev .

# Build with build arguments
docker build --build-arg GO_VERSION=1.25 -t valhalla:dev .
```

## Dependencies

### Viewing Dependencies

```bash
# List all dependencies
go list -m all

# View dependency graph
go mod graph

# Check for available updates
go list -u -m all
```

### Updating Dependencies

```bash
# Update all dependencies to latest minor/patch versions
go get -u ./...

# Update specific dependency
go get github.com/spf13/viper@latest

# Tidy up (remove unused dependencies)
go mod tidy
```

### Vendoring (Optional)

Create a `vendor/` directory with all dependencies:

```bash
# Create vendor directory
go mod vendor

# Build using vendor directory
go build -mod=vendor .
```

This ensures builds are reproducible even if dependency sources change.

## IDE Setup

### Visual Studio Code

1. Install [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go)
2. Configure settings (`.vscode/settings.json`):
   ```json
   {
       "go.useLanguageServer": true,
       "go.lintTool": "golangci-lint",
       "go.lintOnSave": "package",
       "go.formatTool": "goimports",
       "go.buildOnSave": "off"
   }
   ```

### GoLand / IntelliJ IDEA

1. Install Go plugin
2. Open project
3. GoLand will automatically detect `go.mod` and configure the project

### Vim/Neovim

Use [vim-go](https://github.com/fatih/vim-go):

```vim
" In .vimrc
Plug 'fatih/vim-go', { 'do': ':GoUpdateBinaries' }
```

## Debugging

### Delve Debugger

Install Delve:
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Debug a server:
```bash
# Start login server with debugger
dlv debug . -- -type login -config config_login.toml

# In debugger:
(dlv) break main.main
(dlv) continue
(dlv) next
```

### VS Code Debugging

Create `.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Login Server",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}",
            "args": ["-type", "login", "-config", "config_login.toml"]
        },
        {
            "name": "Launch Channel Server",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}",
            "args": ["-type", "channel", "-config", "config_channel_1.toml"]
        }
    ]
}
```

## Performance Profiling

### CPU Profile

```bash
# Build with profiling
go build -o Valhalla .

# Run with CPU profiling
./Valhalla -type channel -config config_channel_1.toml -cpuprofile=cpu.prof

# Analyze profile
go tool pprof cpu.prof
```

### Memory Profile

```bash
# Run with memory profiling
./Valhalla -type channel -config config_channel_1.toml -memprofile=mem.prof

# Analyze
go tool pprof mem.prof
```

### Using pprof HTTP Interface

Add to your code:
```go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

Access profiles at:
- http://localhost:6060/debug/pprof/
- http://localhost:6060/debug/pprof/heap
- http://localhost:6060/debug/pprof/goroutine

## CI/CD Integration

The project uses GitHub Actions for automated builds and releases.

### .goreleaser.yaml

The project uses GoReleaser for creating releases:

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Test release locally (no publish)
goreleaser release --snapshot --clean

# Create actual release (requires git tag)
git tag v1.0.0
goreleaser release --clean
```

## Troubleshooting

### Build Fails with "go: module not found"

```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download
```

### "Command not found" after building

```bash
# Ensure binary has execute permissions (Linux/macOS)
chmod +x Valhalla

# Add to PATH or run with ./
./Valhalla -type login
```

### Import Cycle Errors

Go doesn't allow circular imports. Restructure code to break the cycle, often by:
- Creating a separate package for shared interfaces
- Moving common types to a `types` package
- Using dependency injection

### CGO Errors

If you encounter CGO errors and don't need CGO:

```bash
CGO_ENABLED=0 go build .
```

## Contributing

When contributing code:

1. **Format your code**: `go fmt ./...`
2. **Run tests**: `go test ./...`
3. **Run linter**: `golangci-lint run`
4. **Write tests** for new features
5. **Update documentation** as needed
6. **Follow Go conventions**: Use effective Go practices

## Next Steps

- Set up local environment: [Local.md](Local.md)
- Configure servers: [Configuration.md](Configuration.md)
- Deploy with Docker: [Docker.md](Docker.md)
- Deploy to Kubernetes: [Kubernetes.md](Kubernetes.md)

## Useful Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
- [Go Wiki](https://github.com/golang/go/wiki)
- [Awesome Go](https://awesome-go.com/) - Curated list of Go libraries
