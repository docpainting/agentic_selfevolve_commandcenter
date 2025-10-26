# Fiber v3 Requirements

## ⚠️ Important: Go 1.25+ Required

**Go Fiber v3** requires **Go 1.25 or higher** to function properly. This is a hard requirement due to Fiber v3's use of newer Go features and standard library improvements.

### Why Go 1.25+?

Fiber v3 leverages:
- Enhanced routing performance from Go 1.25
- Improved HTTP/2 and HTTP/3 support
- Better memory management and garbage collection
- New standard library features for WebSocket handling
- Type parameter improvements for generic handlers

### Installation

#### Ubuntu/Debian

```bash
# Remove old Go version (if installed)
sudo rm -rf /usr/local/go

# Download Go 1.25
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz

# Extract to /usr/local
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
# Should output: go version go1.25.0 linux/amd64
```

#### macOS

```bash
# Using Homebrew
brew install go@1.25

# Or download from https://go.dev/dl/
# Then follow the installation wizard
```

#### Windows

Download the installer from: https://go.dev/dl/go1.25.0.windows-amd64.msi

### Verification

After installation, verify Go version:

```bash
go version
```

Expected output:
```
go version go1.25.0 linux/amd64
```

### Common Issues

#### "go.mod requires Go 1.25 or higher"

**Solution:** Update your Go installation to 1.25+

```bash
go version  # Check current version
# If < 1.25, follow installation steps above
```

#### "undefined: fiber.New"

**Solution:** Ensure you're using Fiber v3:

```bash
cd backend
go get github.com/gofiber/fiber/v3@latest
go mod tidy
```

#### Build fails with "unsupported Go version"

**Solution:** Clean module cache and rebuild:

```bash
go clean -modcache
go mod download
go build ./cmd/server
```

### Project Configuration

All `go.mod` files in this project specify:

```go
module agent-workspace/backend

go 1.25

require (
    github.com/gofiber/fiber/v3 v3.0.0
    // ... other dependencies
)
```

### CI/CD Configuration

If using GitHub Actions or other CI/CD:

```yaml
- name: Setup Go
  uses: actions/setup-go@v4
  with:
    go-version: '1.25'
```

### Docker Configuration

If containerizing:

```dockerfile
FROM golang:1.25-alpine

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o server ./cmd/server

CMD ["./server"]
```

### Additional Resources

- **Go 1.25 Release Notes**: https://go.dev/doc/go1.25
- **Fiber v3 Documentation**: https://docs.gofiber.io/
- **Migration Guide**: https://docs.gofiber.io/guide/migration

---

## Summary

✅ **Minimum Go Version**: 1.25.0  
✅ **Recommended**: Latest stable Go 1.25.x  
✅ **Fiber Version**: v3.0.0+  
✅ **Compatibility**: Linux, macOS, Windows  

**Before running this project, ensure `go version` shows 1.25 or higher!**

