# Deployment Guide

This guide covers various deployment strategies for applications built with TheSkyscape DevTools.

## Table of Contents

- [Local Development](#local-development)
- [Docker Deployment](#docker-deployment)
- [Cloud Deployment](#cloud-deployment)
- [Production Considerations](#production-considerations)
- [Monitoring & Logging](#monitoring--logging)
- [Security](#security)

## Local Development

### Basic Setup

1. **Environment Variables**:
   ```bash
   export AUTH_SECRET="development-secret-change-in-production"
   export THEME="corporate"
   export PORT="8080"
   ```

2. **Run Application**:
   ```bash
   go run .
   ```

3. **Development with Hot Reload** (using air):
   ```bash
   go install github.com/cosmtrek/air@latest
   air
   ```

### Development Configuration

Create a `.env.development` file:
```bash
# Development settings
AUTH_SECRET=dev-secret-not-for-production
THEME=corporate
PORT=8080
DEBUG=true
LOG_LEVEL=debug

# Development data directory
INTERNAL_DATA=./dev-data
```

## Docker Deployment

### Single Container

**Dockerfile**:
```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/views ./views

# Create data directory
RUN mkdir -p /data
ENV INTERNAL_DATA=/data

EXPOSE 5000
CMD ["./main"]
```

**Build and Run**:
```bash
docker build -t my-app .
docker run -p 5000:5000 \
  -e AUTH_SECRET="production-secret" \
  -e THEME="corporate" \
  -v app-data:/data \
  my-app
```

### Docker Compose

**docker-compose.yml**:
```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "5000:5000"
    environment:
      - AUTH_SECRET=${AUTH_SECRET}
      - THEME=corporate
      - INTERNAL_DATA=/data
    volumes:
      - app-data:/data
      - ./ssl:/ssl:ro
    env_file:
      - .env.production
    restart: unless-stopped
    
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    restart: unless-stopped
    
  reverse-proxy:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/ssl:ro
    depends_on:
      - app
    restart: unless-stopped

volumes:
  app-data:
  redis-data:
```

**Production Environment** (`.env.production`):
```bash
AUTH_SECRET=your-super-secret-production-key
CONGO_SSL_FULLCHAIN=/ssl/fullchain.pem
CONGO_SSL_PRIVKEY=/ssl/privkey.pem
```

## Cloud Deployment

### DigitalOcean Deployment

Using TheSkyscape DevTools hosting package:

```go
package main

import (
    "os"
    "github.com/The-Skyscape/devtools/pkg/hosting/platforms/digitalocean"
    "github.com/The-Skyscape/devtools/pkg/hosting"
)

func deployToDigitalOcean() error {
    client := digitalocean.Connect(os.Getenv("DIGITAL_OCEAN_API_KEY"))
    
    server := &digitalocean.Server{
        Name:   "my-app-production",
        Size:   "s-2vcpu-4gb",
        Region: "nyc1",
        Image:  "docker-20-04",
    }
    
    deployedServer, err := client.Launch(server,
        hosting.WithFileUpload("./app", "/usr/local/bin/app"),
        hosting.WithFileUpload("./docker-compose.yml", "/root/docker-compose.yml"),
        hosting.WithFileUpload("./.env.production", "/root/.env"),
        hosting.WithSetupScript("docker", "compose", "up", "-d"),
    )
    
    if err != nil {
        return err
    }
    
    // Set up domain
    return deployedServer.Alias("app", "yourdomain.com")
}
```

### Manual DigitalOcean Deployment

1. **Create Droplet**:
   ```bash
   # Using doctl CLI
   doctl compute droplet create my-app \
     --region nyc1 \
     --image docker-20-04 \
     --size s-2vcpu-4gb \
     --ssh-keys your-ssh-key-id
   ```

2. **Deploy Application**:
   ```bash
   # Copy files to server
   scp -r . root@your-server-ip:/app
   
   # SSH and start services
   ssh root@your-server-ip
   cd /app
   docker compose up -d
   ```

### AWS Deployment

```bash
# Using AWS CLI and Docker
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin your-account.dkr.ecr.us-east-1.amazonaws.com

# Build and push
docker build -t my-app .
docker tag my-app:latest your-account.dkr.ecr.us-east-1.amazonaws.com/my-app:latest
docker push your-account.dkr.ecr.us-east-1.amazonaws.com/my-app:latest

# Deploy with ECS or EKS
aws ecs update-service --cluster my-cluster --service my-app-service --force-new-deployment
```

## Production Considerations

### Performance

1. **Build Optimization**:
   ```bash
   # Build with optimizations
   CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .
   ```

2. **Database Configuration**:
   ```go
   // Use WAL mode for better concurrency (default in DevTools)
   // Consider connection pooling for high-traffic apps
   ```

3. **Caching**:
   ```go
   // Add Redis for session storage
   import "github.com/The-Skyscape/devtools/pkg/containers"
   
   redis := &containers.Service{
       Name:  "redis",
       Image: "redis:alpine",
       Ports: map[int]int{6379: 6379},
   }
   ```

### Scaling

1. **Horizontal Scaling**:
   ```yaml
   # docker-compose.yml
   services:
     app:
       deploy:
         replicas: 3
       # ... other config
     
     nginx:
       # Load balancer configuration
   ```

2. **Database Scaling**:
   ```go
   // Consider PostgreSQL for larger applications
   // Use read replicas for read-heavy workloads
   ```

### Health Checks

Add health check endpoint:

```go
// In your main application
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    // Check database connection
    if err := models.DB.Query("SELECT 1").Exec(); err != nil {
        http.Error(w, "Database unhealthy", http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})
```

Docker health check:
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:5000/health || exit 1
```

## Monitoring & Logging

### Application Logging

```go
import "log/slog"

// Structured logging
slog.Info("User logged in", 
    "user_id", user.ID, 
    "ip", r.RemoteAddr)

slog.Error("Database error", 
    "error", err, 
    "query", "SELECT * FROM users")
```

### Metrics Collection

```go
// Add metrics endpoint
import "expvar"

http.Handle("/metrics", expvar.Handler())

// Custom metrics
var requestCount = expvar.NewInt("requests_total")
var errorCount = expvar.NewInt("errors_total")
```

### External Monitoring

1. **Prometheus**:
   ```yaml
   # prometheus.yml
   scrape_configs:
     - job_name: 'my-app'
       static_configs:
         - targets: ['app:5000']
       metrics_path: '/metrics'
   ```

2. **Log Aggregation**:
   ```yaml
   # docker-compose.yml
   logging:
     driver: "json-file"
     options:
       max-size: "10m"
       max-file: "3"
   ```

## Security

### SSL/TLS Setup

1. **Let's Encrypt with Certbot**:
   ```bash
   # Install certbot
   sudo apt-get install certbot
   
   # Get certificate
   sudo certbot certonly --standalone -d yourdomain.com
   
   # Certificates will be in /etc/letsencrypt/live/yourdomain.com/
   ```

2. **Environment Variables**:
   ```bash
   export CONGO_SSL_FULLCHAIN="/etc/letsencrypt/live/yourdomain.com/fullchain.pem"
   export CONGO_SSL_PRIVKEY="/etc/letsencrypt/live/yourdomain.com/privkey.pem"
   ```

### Firewall Configuration

```bash
# UFW (Ubuntu)
sudo ufw allow ssh
sudo ufw allow http
sudo ufw allow https
sudo ufw enable

# Or specific ports
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

### Security Headers

Add security middleware:

```go
func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        next.ServeHTTP(w, r)
    })
}
```

### Secrets Management

1. **Environment Variables**:
   ```bash
   # Never commit secrets to git
   # Use environment variables or secret management services
   ```

2. **Docker Secrets**:
   ```yaml
   # docker-compose.yml
   secrets:
     auth_secret:
       external: true
   
   services:
     app:
       secrets:
         - auth_secret
   ```

### Backup Strategy

1. **Database Backup**:
   ```bash
   # Backup SQLite database
   sqlite3 /data/app.db ".backup /backups/app-$(date +%Y%m%d-%H%M%S).db"
   ```

2. **Automated Backups**:
   ```bash
   # Cron job for daily backups
   0 2 * * * /usr/local/bin/backup-script.sh
   ```

3. **File Backups**:
   ```bash
   # Backup uploaded files and data
   rsync -avz /data/ backup-server:/backups/app-data/
   ```

### Maintenance

1. **Update Strategy**:
   ```bash
   # Blue-green deployment
   docker compose -f docker-compose.blue.yml up -d
   # Switch traffic
   docker compose -f docker-compose.green.yml down
   ```

2. **Database Migrations**:
   ```go
   // Run migrations on startup
   // Use migration tools for schema changes
   ```

## Troubleshooting

### Common Issues

1. **Port Conflicts**:
   ```bash
   # Check what's using a port
   sudo netstat -tulpn | grep :5000
   sudo lsof -i :5000
   ```

2. **Database Issues**:
   ```bash
   # Check database file permissions
   ls -la /data/app.db
   
   # Check disk space
   df -h
   ```

3. **SSL Certificate Issues**:
   ```bash
   # Check certificate validity
   openssl x509 -in /ssl/fullchain.pem -text -noout
   
   # Test SSL connection
   openssl s_client -connect yourdomain.com:443
   ```

### Debug Mode

Enable debug logging:

```bash
export DEBUG=true
export LOG_LEVEL=debug
```

### Performance Profiling

```go
import _ "net/http/pprof"

// Access profiling at /debug/pprof/
```

For more specific deployment scenarios or questions, refer to the [API Documentation](api.md) or check the [Examples](examples/) directory.