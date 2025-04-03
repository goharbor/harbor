#!/bin/bash

# Detect platform
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
  PLATFORM="linux/amd64"
elif [[ "$ARCH" == "aarch64" ]]; then
  PLATFORM="linux/arm64"
else
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

# Generate docker-compose.override.yml
cat > docker-compose.override.yml <<EOF
version: '2.3'

services:
  log:
    image: ranichowdary/log-harbor:latest
    platform: $PLATFORM

  registry:
    image: ranichowdary/registry-harbor:latest
    platform: $PLATFORM

  registryctl:
    image: ranichowdary/registryctl-harbor:latest
    platform: $PLATFORM

  postgresql:
    image: ranichowdary/db-harbor:latest
    platform: $PLATFORM

  core:
    image: ranichowdary/core-harbor:latest
    platform: $PLATFORM

  portal:
    image: ranichowdary/portal-harbor:latest
    platform: $PLATFORM

  jobservice:
    image: ranichowdary/jobservice-harbor:latest
    platform: $PLATFORM

  redis:
    image: ranichowdary/redis-harbor:latest
    platform: $PLATFORM

  proxy:
    image: ranichowdary/nginx-harbor:latest
    platform: $PLATFORM
EOF

echo "docker-compose.override.yml generated for $PLATFORM"
