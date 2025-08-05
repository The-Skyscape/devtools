#!/bin/bash

# Simple deployment script based on Featherweight pattern
# Parameters interpolated by Go
DOMAIN="%s"
EMAIL="%s"
API_TOKEN="%s"
REDEPLOY="%s"
AUTH_SECRET="%s"

echo "Starting deployment..."

# Configuration
BINARY_PATH="/root/app"
CONTAINER_NAME="sky-app"
IMAGE_NAME="skyscape:latest"
CONTAINER_BINARY="/app"

# Build Docker image if Dockerfile exists
if [ -f "/root/Dockerfile" ]; then
    echo "Building Docker image..."
    docker build -t "$IMAGE_NAME" /root/
fi

# Stop and remove existing container
echo "Cleaning up existing container..."
docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true

# Create new container
echo "Creating new container..."
CONTAINER_ID=$(docker create \
  --name "$CONTAINER_NAME" \
  --entrypoint "$CONTAINER_BINARY" \
  --network host \
  --privileged \
  --restart unless-stopped \
  -v "/root/.skyscape:/root/.skyscape" \
  -v "/var/run/docker.sock:/var/run/docker.sock" \
  -e PORT=80 \
  -e THEME=corporate \
  -e AUTH_SECRET="$AUTH_SECRET" \
  "$IMAGE_NAME")

# Copy binary into container
echo "Copying application binary..."
docker cp "$BINARY_PATH" "${CONTAINER_ID}:${CONTAINER_BINARY}"

# Make binary executable
docker exec "$CONTAINER_ID" chmod +x "$CONTAINER_BINARY"

# Copy SSL certificates if they exist
if [ -f "/root/fullchain.pem" ] && [ -f "/root/privkey.pem" ]; then
    echo "Copying SSL certificates..."
    docker cp /root/fullchain.pem "${CONTAINER_ID}:/root/fullchain.pem"
    docker cp /root/privkey.pem "${CONTAINER_ID}:/root/privkey.pem"
fi

# Start container
echo "Starting container..."
docker start "$CONTAINER_NAME"

# Wait for service to be ready
echo "Waiting for service to start..."
sleep 5

# Check if service is running
if docker ps | grep -q "$CONTAINER_NAME"; then
    echo "✅ Deployment successful! Container '$CONTAINER_NAME' is running"
    
    # Show container logs
    echo "Recent logs:"
    docker logs "$CONTAINER_NAME" --tail 10
else
    echo "❌ Deployment failed! Container is not running"
    echo "Container logs:"
    docker logs "$CONTAINER_NAME" --tail 20
    exit 1
fi

# Handle SSL certificate generation separately if needed
if [ -n "$DOMAIN" ] && [ "$DOMAIN" != "null" ] && [ ! -f "/root/fullchain.pem" ]; then
    echo "Note: SSL certificates not found. You may need to set them up separately."
fi

echo "Deployment complete!"