#!/bin/bash

# Skyscape Application Deployment Script
# This script handles the complete deployment process in a single SSH session
# to avoid connection throttling and improve reliability.

set -e  # Exit on any error

# Script parameters (interpolated by Go)
DOMAIN="%s"
EMAIL="%s"
API_TOKEN="%s"
REDEPLOY="%s"
AUTH_SECRET="%s"

# Color codes for better output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to wait for service to be ready
wait_for_service() {
    local max_attempts=30
    local attempt=1
    
    log_info "Waiting for service to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s http://localhost:80 >/dev/null 2>&1; then
            log_success "Service is ready!"
            return 0
        fi
        
        log_info "Attempt $attempt/$max_attempts - Service not ready yet, waiting..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_error "Service failed to start within expected time"
    return 1
}

# Main deployment function
deploy_application() {
    log_info "Starting Skyscape application deployment..."
    
    # Step 1: Stop and remove any existing container
    log_info "Step 1: Cleaning up existing containers..."
    if docker ps -a --format "table {{.Names}}" | grep -q "^sky-app$"; then
        log_info "Stopping existing sky-app container..."
        docker stop sky-app 2>/dev/null || true
        
        log_info "Removing existing sky-app container..."
        docker rm -f sky-app 2>/dev/null || true
        log_success "Existing container cleaned up"
    else
        log_info "No existing container found"
    fi
    
    # Step 2: Create necessary directories
    log_info "Step 2: Creating necessary directories..."
    mkdir -p /root/.skyscape || {
        log_error "Failed to create /root/.skyscape directory"
        exit 1
    }
    
    # Ensure proper permissions
    chmod 755 /root/.skyscape
    log_success "Directories created and configured"
    
    # Step 3: Build Docker image
    log_info "Step 3: Building Docker image..."
    if ! docker build -t skyscape:latest /root/ 2>&1; then
        log_error "Failed to build Docker image"
        exit 1
    fi
    log_success "Docker image built successfully"
    
    # Step 4: Create and start the container
    log_info "Step 4: Creating and starting container..."
    
    # Create the container with proper configuration
    DOCKER_CREATE_CMD="docker create --name sky-app --network host --privileged --restart unless-stopped --entrypoint /app"
    DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD -v /root/.skyscape:/root/.skyscape"
    DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD -v /var/run/docker.sock:/var/run/docker.sock"
    DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD -e PORT=80"
    DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD -e THEME=corporate"
    
    # Add AUTH_SECRET if provided
    if [ -n "$AUTH_SECRET" ]; then
        DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD -e AUTH_SECRET=$AUTH_SECRET"
    fi
    
    DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD skyscape:latest"
    
    if ! eval $DOCKER_CREATE_CMD 2>&1; then
        log_error "Failed to create container"
        exit 1
    fi
    
    # Copy the application binary into the container
    if ! docker cp /root/app sky-app:/app 2>&1; then
        log_error "Failed to copy application binary"
        exit 1
    fi
    
    # Make the binary executable using docker cp trick (copy it back with execute permissions)
    # First, create a temp executable version
    cp /root/app /root/app-exec
    chmod +x /root/app-exec
    docker cp /root/app-exec sky-app:/app
    rm -f /root/app-exec
    
    # Start the container
    if ! docker start sky-app 2>&1; then
        log_error "Failed to start container"
        exit 1
    fi
    
    log_success "Container created and started successfully"
    
    # Step 5: Wait for service to be ready
    if ! wait_for_service; then
        log_error "Service failed to start properly"
        # Show container logs for debugging
        log_info "Container logs:"
        docker logs sky-app --tail 20
        exit 1
    fi
    
    # Step 6: Configure SSL if domain is provided
    if [ -n "$DOMAIN" ] && [ "$DOMAIN" != "null" ] && [ "$DOMAIN" != "" ]; then
        log_info "Step 6: Configuring SSL certificates for domain: $DOMAIN"
        configure_ssl "$DOMAIN" "$EMAIL" "$API_TOKEN"
    else
        log_info "Step 6: Skipping SSL configuration (no domain provided)"
    fi
    
    log_success "🎉 Deployment completed successfully!"
    log_info "Application is running at:"
    log_info "  - HTTP: http://$(curl -s ifconfig.me || echo 'SERVER_IP')"
    if [ -n "$DOMAIN" ] && [ "$DOMAIN" != "null" ] && [ "$DOMAIN" != "" ]; then
        log_info "  - HTTPS: https://$DOMAIN"
    fi
}

# SSL configuration function
configure_ssl() {
    local domain="$1"
    local email="$2"
    local api_token="$3"
    
    log_info "Installing certbot and DigitalOcean plugin..."
    
    # Update package list first
    apt-get update -qq
    
    # Install certbot if not already installed
    if ! command_exists certbot; then
        log_info "Installing certbot and python3-certbot-dns-digitalocean..."
        if ! apt-get install -y certbot python3-certbot-dns-digitalocean; then
            log_error "Failed to install certbot"
            return 1
        fi
    else
        log_info "Certbot already installed, checking for DigitalOcean plugin..."
        if ! apt-get install -y python3-certbot-dns-digitalocean; then
            log_error "Failed to install DigitalOcean plugin"
            return 1
        fi
    fi
    
    # Create credentials file
    log_info "Creating DigitalOcean credentials file..."
    echo "dns_digitalocean_token=$api_token" > /root/certbot-creds.ini
    chmod 600 /root/certbot-creds.ini
    
    # Generate SSL certificate
    log_info "Generating SSL certificate for $domain..."
    if certbot certonly \
        --dns-digitalocean \
        --dns-digitalocean-credentials /root/certbot-creds.ini \
        -d "$domain" \
        --non-interactive \
        --expand \
        --agree-tos \
        --email "$email" 2>&1; then
        
        log_success "SSL certificate generated successfully"
        
        # Copy certificates to accessible location
        log_info "Copying SSL certificates..."
        if [ -f "/etc/letsencrypt/live/$domain/fullchain.pem" ] && [ -f "/etc/letsencrypt/live/$domain/privkey.pem" ]; then
            cp "/etc/letsencrypt/live/$domain/fullchain.pem" /root/fullchain.pem
            cp "/etc/letsencrypt/live/$domain/privkey.pem" /root/privkey.pem
            
            # Stop the current container
            log_info "Restarting container with SSL certificates..."
            docker stop sky-app
            docker rm -f sky-app
            
            # Recreate container with SSL support
            DOCKER_CREATE_CMD="docker create --name sky-app --network host --privileged --restart unless-stopped --entrypoint /app"
            DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD -v /root/.skyscape:/root/.skyscape"
            DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD -v /var/run/docker.sock:/var/run/docker.sock"
            DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD -e PORT=80"
            DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD -e THEME=corporate"
            DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD -e CONGO_SSL_FULLCHAIN=/root/fullchain.pem"
            DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD -e CONGO_SSL_PRIVKEY=/root/privkey.pem"
            
            # Add AUTH_SECRET if provided
            if [ -n "$AUTH_SECRET" ]; then
                DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD -e AUTH_SECRET=$AUTH_SECRET"
            fi
            
            DOCKER_CREATE_CMD="$DOCKER_CREATE_CMD skyscape:latest"
            
            if eval $DOCKER_CREATE_CMD 2>&1; then
                
                # Copy files into the new container
                docker cp /root/app sky-app:/app
                docker cp /root/fullchain.pem sky-app:/root/fullchain.pem 2>/dev/null || log_warning "Could not copy fullchain.pem"
                docker cp /root/privkey.pem sky-app:/root/privkey.pem 2>/dev/null || log_warning "Could not copy privkey.pem"
                
                # Make binary executable and start
                docker exec sky-app chmod +x /app
                docker start sky-app
                
                log_success "Container restarted with SSL support"
                
                # Wait for service to be ready again
                wait_for_service
                
            else
                log_error "Failed to recreate container with SSL"
                return 1
            fi
        else
            log_error "SSL certificate files not found"
            return 1
        fi
        
        # Clean up credentials file
        rm -f /root/certbot-creds.ini
        
    else
        log_error "Failed to generate SSL certificate"
        rm -f /root/certbot-creds.ini
        return 1
    fi
}

# Health check function
health_check() {
    log_info "Performing health check..."
    
    # Check if container is running
    if ! docker ps | grep -q "sky-app"; then
        log_error "Container is not running"
        return 1
    fi
    
    # Check if service responds
    if ! curl -f -s http://localhost:80 >/dev/null; then
        log_error "Service is not responding"
        return 1
    fi
    
    log_success "Health check passed"
    return 0
}

# Cleanup function for error handling
cleanup_on_error() {
    log_warning "Cleaning up due to error..."
    docker stop sky-app 2>/dev/null || true
    docker rm -f sky-app 2>/dev/null || true
    rm -f /root/certbot-creds.ini 2>/dev/null || true
}

# Set up error handling
trap cleanup_on_error ERR

# Main execution
main() {
    log_info "Skyscape Deployment Script v1.0"
    log_info "================================"
    
    # Validate Docker is available
    if ! command_exists docker; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    # Check if Docker daemon is running
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker daemon is not running"
        exit 1
    fi
    
    # Check if required files exist
    if [ ! -f "/root/app" ]; then
        log_error "Application binary not found at /root/app"
        exit 1
    fi
    
    if [ ! -f "/root/Dockerfile" ]; then
        log_error "Dockerfile not found at /root/Dockerfile"
        exit 1
    fi
    
    # Execute deployment
    deploy_application
    
    # Final health check
    if health_check; then
        log_success "🚀 Deployment completed successfully and service is healthy!"
        exit 0
    else
        log_error "Deployment completed but health check failed"
        exit 1
    fi
}

# Execute main function
main "$@"