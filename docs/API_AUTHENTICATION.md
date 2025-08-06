# ClamAV API Configuration Examples

## Environment Variables for API Key Authentication

### Enable API Key Authentication

```bash
# Set the API key (required for authentication)
export AUTH_API_KEY="your-secure-api-key-here"

# Optional: Set custom header name (default: X-API-Key)
export AUTH_API_KEY_HEADER="Authorization"
```

### Disable API Key Authentication (Default)

```bash
# Leave AUTH_API_KEY empty or unset to disable authentication
unset AUTH_API_KEY
# OR
export AUTH_API_KEY=""
```

## Usage Examples

### With Authentication Enabled

```bash
# Health check (always public, no authentication required)
curl -X GET http://localhost:8080/rest/v1/ping

# Protected endpoints (require API key)
curl -X GET \
  -H "X-API-Key: your-secure-api-key-here" \
  http://localhost:8080/rest/v1/version

curl -X POST \
  -H "X-API-Key: your-secure-api-key-here" \
  -F "file=@test-file.txt" \
  http://localhost:8080/rest/v1/scan
```

### With Custom Authorization Header

```bash
# Set custom header name
export AUTH_API_KEY="your-secure-api-key-here"
export AUTH_API_KEY_HEADER="Authorization"

# Use custom header
curl -X GET \
  -H "Authorization: your-secure-api-key-here" \
  http://localhost:8080/rest/v1/version
```

### Production Security Recommendations

1. **Generate Strong API Keys**:

   ```bash
   # Generate a secure random API key
   openssl rand -hex 32
   # OR
   python3 -c "import secrets; print(secrets.token_urlsafe(32))"
   ```

2. **Use Environment Variables**:
   - Never hardcode API keys in configuration files
   - Use environment variables or secret management systems
   - Rotate API keys regularly

3. **HTTPS Only**:
   - Always use HTTPS in production
   - API keys in headers are transmitted with every request

4. **Multiple API Keys** (Future Enhancement):
   - Consider implementing multiple API keys with different permissions
   - API key rotation capabilities
   - Rate limiting per API key

## Public Endpoints (Always Accessible)

The following endpoints remain accessible without authentication:

- `/rest/v1/ping` - Health check
- `/health` - Alternative health check
- `/readiness` - Kubernetes readiness probe
- `/liveness` - Kubernetes liveness probe

## Protected Endpoints (Require Authentication When Enabled)

- `/rest/v1/version` - ClamAV version information
- `/rest/v1/stats` - ClamAV statistics
- `/rest/v1/versioncommands` - ClamAV version commands
- `/rest/v1/scan` - File scanning
- `/rest/v1/reload` - Reload ClamAV configuration
- `/rest/v1/shutdown` - Shutdown ClamAV daemon

## Error Responses

### Missing API Key

```json
{
  "status": "error",
  "msg": "API key required"
}
```

HTTP Status: `401 Unauthorized`
Headers: `WWW-Authenticate: API-Key`

### Invalid API Key

```json
{
  "status": "error",
  "msg": "Invalid API key"
}

```
HTTP Status: `401 Unauthorized`
Headers: `WWW-Authenticate: API-Key`
