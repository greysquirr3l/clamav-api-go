# ClamAV API Go 🛡️

![Go](https://img.shields.io/badge/go-1.23+-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Docker](https://img.shields.io/badge/docker-ready-blue)
![Security](https://img.shields.io/badge/security-authenticated-red)

Production-ready REST API wrapper for [ClamAV](http://www.clamav.net/) antivirus scanning with
enterprise-grade security, authentication, and comprehensive virus detection capabilities.

## 🚀 Features

- **🔒 Optional API Key Authentication** - Secure your endpoints with industry-standard authentication
- **🦠 Comprehensive Virus Scanning** - Full ClamAV protocol support with optimized configurations
- **🔄 Remote Virus Database Updates** - FreshClam API endpoint for automated definition management
- **📊 Real-time Monitoring** - Health checks, statistics, and operational status endpoints
- **🐳 Container Ready** - Production-optimized Docker deployment with multi-container architecture
- **⚡ High Performance** - Optimized memory allocation supporting 1.6GB+ virus databases
- **📝 Structured Logging** - JSON/Console logging with request correlation and observability
- **🏗️ Clean Architecture** - Interface-driven design with comprehensive test coverage

## 🎯 Use Cases

- **Security Gateways** - Scan file uploads in web applications and APIs
- **Email Security** - Integrate antivirus scanning into email processing pipelines
- **Content Management** - Protect document management systems and file repositories
- **CI/CD Security** - Scan artifacts and dependencies in automated build pipelines
- **Microservices** - Add virus scanning capabilities to distributed applications

## 📋 Requirements

- **Go**: 1.23.0 or higher
- **ClamAV**: Latest stable version recommended
- **Docker**: For containerized deployment
- **Memory**: 4GB+ RAM recommended for production (virus databases require 1.6GB+)

## 🚀 Quick Start

### Docker Deployment (Recommended)

#### Production with Security & Optimization

```bash
# Production deployment with optimized configuration and security
docker compose -f docker-compose.optimized.yaml up -d
```

#### Development

```bash
# Basic development setup
docker compose up -d
```

The API server will be available at <http://127.0.0.1:8888>

### Binary Installation

```bash
# Download the latest release
wget https://github.com/lescactus/clamav-api-go/releases/latest/download/clamav-api-go

# Make executable and run
chmod +x clamav-api-go && ./clamav-api-go
```

### Building from Source

```bash
# Clone and build
git clone https://github.com/lescactus/clamav-api-go.git
cd clamav-api-go && go build -o bin/clamav-api-go && ./bin/clamav-api-go
```

## 📡 API Endpoints

### Health & Monitoring

| Method | Endpoint | Description | Authentication |
|--------|----------|-------------|----------------|
| `GET` | `/rest/v1/ping` | Health check and ClamAV connectivity | Public |
| `GET` | `/rest/v1/version` | ClamAV version information | Protected |
| `GET` | `/rest/v1/stats` | ClamAV daemon statistics | Protected |
| `GET` | `/rest/v1/versioncommands` | Available ClamAV commands | Protected |

### Virus Scanning

| Method | Endpoint | Description | Authentication |
|--------|----------|-------------|----------------|
| `POST` | `/rest/v1/scan` | Scan uploaded files for viruses | Protected |

### Management Operations

| Method | Endpoint | Description | Authentication |
|--------|----------|-------------|----------------|
| `POST` | `/rest/v1/reload` | Reload ClamAV configuration | Protected |
| `POST` | `/rest/v1/shutdown` | Shutdown ClamAV daemon | Protected |
| `POST` | `/rest/v1/freshclam` | Update virus definitions | Protected |

## 🔒 API Authentication

The ClamAV API supports optional API key authentication for production security. When enabled,
all protected endpoints require a valid API key in the request header.

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AUTH_API_KEY` | `""` (disabled) | API key for authentication. If empty, authentication is disabled |
| `AUTH_API_KEY_HEADER` | `X-API-Key` | Header name for API key authentication |

### Security Features

- **Optional Authentication**: Can be enabled/disabled via configuration
- **Public Health Endpoints**: Health checks remain accessible without authentication
- **Secure Key Comparison**: Uses constant-time comparison to prevent timing attacks
- **Comprehensive Logging**: Failed authentication attempts are logged with client details
- **Flexible Headers**: Customizable API key header name

### Usage Examples

#### Generate Secure API Key

```bash
# Generate a cryptographically secure API key
export AUTH_API_KEY=$(openssl rand -hex 32)
echo "Generated API Key: $AUTH_API_KEY"
```

#### Making Authenticated Requests

```bash
# Health check (always public)
curl -X GET http://localhost:8888/rest/v1/ping

# Protected endpoints (require API key)
curl -X GET \
  -H "X-API-Key: your-api-key-here" \
  http://localhost:8888/rest/v1/version

# File scanning with authentication
curl -X POST \
  -H "X-API-Key: your-api-key-here" \
  -F "file=@suspicious-file.txt" \
  http://localhost:8888/rest/v1/scan
```

## ⚙️ Configuration

The application uses [Viper](https://github.com/spf13/viper) for 12-factor compliant configuration
management. Configuration can be provided via environment variables, config files, or command-line
flags.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_ADDR` | `:8888` | Server listening address (host:port) |
| `SERVER_READ_TIMEOUT` | `30s` | Maximum duration for reading requests |
| `SERVER_WRITE_TIMEOUT` | `30s` | Maximum duration for writing responses |
| `SERVER_MAX_REQUEST_SIZE` | `104857600` | Maximum request size in bytes (100MB) |
| `CLAMAV_ADDR` | `127.0.0.1:3310` | ClamAV daemon address |
| `CLAMAV_NETWORK` | `tcp` | Network type for ClamAV connection |
| `CLAMAV_TIMEOUT` | `30s` | ClamAV connection timeout |
| `LOGGER_LOG_LEVEL` | `info` | Log level (trace, debug, info, warn, error, fatal, panic) |
| `LOGGER_FORMAT` | `json` | Log format (json or console) |
| `AUTH_API_KEY` | `""` | API key for authentication (empty = disabled) |
| `AUTH_API_KEY_HEADER` | `X-API-Key` | Header name for API key |

### Configuration Files

Supported formats: `config.json`, `config.yaml`, `config.env`

#### Example config.yaml

```yaml
server_addr: ":8888"
server_read_timeout: "30s"
server_write_timeout: "30s"
server_max_request_size: 104857600
clamav_addr: "127.0.0.1:3310"
clamav_network: "tcp"
clamav_timeout: "30s"
logger_log_level: "info"
logger_format: "json"
auth_api_key: "your-secure-api-key-here"
auth_api_key_header: "X-API-Key"
```

## 🐳 Docker Deployment

### Production Configuration

The optimized Docker Compose configuration provides enterprise-ready deployment with:

- **Memory Management**: 4GB allocation for ClamAV, 512MB for API gateway
- **Security Architecture**: ClamAV daemon isolated from external access
- **Resource Limits**: CPU and memory constraints for stable operation
- **Health Checks**: Comprehensive service monitoring
- **Optimized ClamAV**: Enhanced detection capabilities based on Arch Linux recommendations

```bash
# Deploy production stack
docker compose -f docker-compose.optimized.yaml up -d

# Monitor services
docker compose -f docker-compose.optimized.yaml logs -f

# Scale API gateway (if needed)
docker compose -f docker-compose.optimized.yaml up -d --scale clamav-api=3
```

### Development Configuration

```bash
# Quick development setup
docker compose up -d

# View logs
docker compose logs -f clamav-api
```

## 📝 Usage Examples

### File Scanning

#### Clean File

```bash
curl -X POST \
  -F "file=@clean-document.pdf" \
  http://localhost:8888/rest/v1/scan | jq

# Response
{
  "status": "noerror",
  "msg": "stream: OK",
  "signature": "",
  "virus_found": false
}
```

#### Infected File (EICAR Test)

```bash
curl -X POST \
  -F "file=@eicar.txt" \
  http://localhost:8888/rest/v1/scan | jq

# Response
{
  "status": "error",
  "msg": "stream: Win.Test.EICAR_HDB-1 FOUND",
  "signature": "Win.Test.EICAR_HDB-1",
  "virus_found": true
}
```

### System Information

#### Health Check

```bash
curl http://localhost:8888/rest/v1/ping | jq

# Response
{
  "ping": "PONG"
}
```

#### ClamAV Statistics

```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:8888/rest/v1/stats | jq

# Response
{
  "stats": "POOLS: 1\nSTATE: VALID PRIMARY\nTHREADS: live 1  idle 0 max 12 idle-timeout 30\n..."
}
```

### Virus Definition Updates

```bash
curl -X POST \
  -H "X-API-Key: your-api-key" \
  http://localhost:8888/rest/v1/freshclam | jq

# Response
{
  "status": "success",
  "message": "Virus definitions updated successfully",
  "timestamp": "2025-08-06T10:30:00Z",
  "output": "ClamAV update process started at Tue Aug  6 10:30:00 2025\n..."
}
```

## 🛠️ Development

### Prerequisites

- Go 1.23.0+
- ClamAV daemon running on localhost:3310
- Make (optional, for build automation)

### Local Development

```bash
# Clone repository
git clone https://github.com/lescactus/clamav-api-go.git
cd clamav-api-go

# Install dependencies
go mod download

# Run tests
make test

# Build and run
make build && ./bin/clamav-api-go
```

### Testing

#### Unit Tests

```bash
# Run all unit tests
go test -v ./...

# Run with coverage
go test -v -cover -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

#### End-to-End Tests

```bash
# Start services
docker compose up -d --wait

# Run E2E tests with Venom
venom run -vv e2e/venom.e2e.yaml

# Custom target
venom run -vv --var=baseuri=https://api.example.com e2e/venom.e2e.yaml
```

### Build Targets

```bash
# Available make targets
make help

# Common operations
make build          # Build binary
make test           # Run tests
make lint           # Run linters
make docker-build   # Build Docker image
make clean          # Clean build artifacts
```

## 🚢 Production Deployment

### Memory Requirements

Based on Arch Linux ClamAV recommendations:

- **ClamAV Daemon**: 1.6GB+ for virus definitions, 3.2GB peak during updates
- **API Gateway**: 256MB-512MB for Go application
- **Total Recommended**: 4GB+ RAM for stable production operation

### Security Best Practices

1. **Enable API Key Authentication**

   ```bash
   export AUTH_API_KEY=$(openssl rand -hex 32)
   ```

2. **Use HTTPS in Production**

   ```bash
   # Deploy behind reverse proxy (nginx, Caddy, etc.)
   # Configure TLS termination and rate limiting
   ```

3. **Resource Limits**

   ```yaml
   # Docker Compose resource constraints
   deploy:
     resources:
       limits:
         memory: 4G
         cpus: '2.0'
   ```

4. **Network Security**

   ```yaml
   # Isolate ClamAV daemon from external access
   expose:
     - "3310"  # Internal only, not ports
   ```

### Monitoring & Observability

#### Health Checks

```bash
# Kubernetes readiness probe
curl -f http://localhost:8888/rest/v1/ping || exit 1

# Docker health check
curl --fail http://localhost:8888/rest/v1/ping
```

#### Metrics Collection

```bash
# Prometheus metrics endpoint (if enabled)
curl http://localhost:8888/metrics

# Structured logs for analysis
docker logs clamav-api-gateway | jq '.level="error"'
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run the test suite (`make test`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Code Standards

- Follow Go conventions and `gofmt` formatting
- Add unit tests for new functionality
- Update documentation for API changes
- Use conventional commit messages

## 📚 Documentation

- [API Authentication Guide](docs/API_AUTHENTICATION.md)
- [Security Policy](.github/SECURITY.md)
- [ClamAV Protocol Reference](http://linux.die.net/man/8/clamd)
- [Docker Deployment Guide](docs/DOCKER_DEPLOYMENT.md)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [ClamAV Project](https://www.clamav.net/) for the excellent antivirus engine
- [Arch Linux ClamAV Wiki](https://wiki.archlinux.org/title/ClamAV) for optimization recommendations
- Go community for amazing libraries and tools

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/lescactus/clamav-api-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/lescactus/clamav-api-go/discussions)
- **Security**: See [SECURITY.md](.github/SECURITY.md) for reporting vulnerabilities

---

Built with ❤️ using Go and ClamAV
