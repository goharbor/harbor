# E2E Test Setup: Artifact Detail Additions (Dockerfile Tab)

## Overview

This document describes how to prepare test images for the Artifact Detail Additions E2E tests, specifically for the new Dockerfile display feature.

## Test Images Required

### 1. Image WITH Dockerfile Label
**Project:** `test-dockerfile-with-label`
**Repository:** `test-image`
**Tag:** `latest`

#### Build Command
```bash
# Create a simple test Dockerfile
cat > Dockerfile << 'EOF'
FROM ubuntu:22.04
RUN apt-get update
RUN apt-get install -y curl
COPY test.txt /app/
LABEL org.opencontainers.image.authors="Test User"
LABEL org.opencontainers.image.source="FROM ubuntu:22.04\nRUN apt-get update\nRUN apt-get install -y curl\nCOPY test.txt /app/\nLABEL org.opencontainers.image.authors=\"Test User\""
EOF

# Create a test file to copy
echo "test content" > test.txt

# Build image (adjust HARBOR_REGISTRY as needed)
export HARBOR_REGISTRY="harbor.example.com"
export DOCKER_BUILDKIT=1

docker build \
  --label "org.opencontainers.image.source=$(cat Dockerfile | sed 's/"/\\"/g')" \
  -t ${HARBOR_REGISTRY}/test-dockerfile-with-label/test-image:latest \
  .

docker push ${HARBOR_REGISTRY}/test-dockerfile-with-label/test-image:latest
```

#### Docker Compose Alternative
```yaml
version: '3.8'
services:
  build-with-label:
    build:
      context: .
      dockerfile: Dockerfile.with-label
    image: harbor.example.com/test-dockerfile-with-label/test-image:latest
    labels:
      org.opencontainers.image.source: "FROM ubuntu:22.04\nRUN apt-get update\nRUN apt-get install -y curl"
```

### 2. Image WITHOUT Dockerfile Label
**Project:** `test-dockerfile-without-label`
**Repository:** `test-image`
**Tag:** `latest`

#### Build Command
```bash
# Create a simple test Dockerfile (same content, no Dockerfile label)
cat > Dockerfile << 'EOF'
FROM ubuntu:22.04
RUN apt-get update
RUN apt-get install -y curl
COPY test.txt /app/
EOF

# Create a test file to copy
echo "test content" > test.txt

# Build image WITHOUT Dockerfile label
export HARBOR_REGISTRY="harbor.example.com"
export DOCKER_BUILDKIT=1

docker build \
  -t ${HARBOR_REGISTRY}/test-dockerfile-without-label/test-image:latest \
  .

docker push ${HARBOR_REGISTRY}/test-dockerfile-without-label/test-image:latest
```

## Test Automation Setup

### Pre-Test Hook
Add to your test setup script to automatically build and push test images:

```bash
#!/bin/bash
set -e

HARBOR_URL="${HARBOR_URL:-https://harbor.example.com}"
HARBOR_REGISTRY="${HARBOR_REGISTRY:-harbor.example.com}"
TEST_USER="${HARBOR_USER:-admin}"
TEST_PASSWORD="${HARBOR_PASSWORD:-Harbor12345}"

# Login to harbor
echo "$TEST_PASSWORD" | docker login -u "$TEST_USER" --password-stdin "$HARBOR_REGISTRY"

# Create test directories
mkdir -p /tmp/harbor-test-images/with-label
mkdir -p /tmp/harbor-test-images/without-label

# Build image WITH Dockerfile label
cd /tmp/harbor-test-images/with-label
cat > Dockerfile << 'EOF'
FROM ubuntu:22.04
RUN apt-get update
RUN apt-get install -y curl
COPY test.txt /app/
LABEL org.opencontainers.image.authors="Harbor Test"
EOF

echo "test content" > test.txt
DOCKERFILE_CONTENT=$(cat Dockerfile | base64 | tr -d '\n')

docker build \
  --label "org.opencontainers.image.source=$(cat Dockerfile)" \
  -t "$HARBOR_REGISTRY/test-dockerfile-with-label/test-image:latest" \
  .

docker push "$HARBOR_REGISTRY/test-dockerfile-with-label/test-image:latest"

# Build image WITHOUT Dockerfile label
cd /tmp/harbor-test-images/without-label
cat > Dockerfile << 'EOF'
FROM ubuntu:22.04
RUN apt-get update
RUN apt-get install -y curl
COPY test.txt /app/
EOF

echo "test content" > test.txt

docker build \
  -t "$HARBOR_REGISTRY/test-dockerfile-without-label/test-image:latest" \
  .

docker push "$HARBOR_REGISTRY/test-dockerfile-without-label/test-image:latest"

echo "Test images created successfully"
```

### Robot Framework Setup
Add to `tests/resources/Util.robot`:

```robot
*** Keywords ***

Push Test Image With Dockerfile Label
    [Arguments]    ${project}    ${repo}    ${tag}
    [Documentation]    Build and push test image with Dockerfile label
    ${cmd}=    Set Variable    \
        mkdir -p /tmp/test-build && \
        cd /tmp/test-build && \
        cat > Dockerfile << 'EOF'\nFROM ubuntu:22.04\nRUN apt-get update\nRUN apt-get install -y curl\nEOF\n && \
        docker build --label "org.opencontainers.image.source=$(cat Dockerfile)" \
            -t ${LOCAL_REGISTRY}/${project}/${repo}:${tag} . && \
        docker push ${LOCAL_REGISTRY}/${project}/${repo}:${tag}
    Execute Command In Host    ${cmd}

Push Test Image Without Dockerfile Label
    [Arguments]    ${project}    ${repo}    ${tag}
    [Documentation]    Build and push test image without Dockerfile label
    ${cmd}=    Set Variable    \
        mkdir -p /tmp/test-build && \
        cd /tmp/test-build && \
        cat > Dockerfile << 'EOF'\nFROM ubuntu:22.04\nRUN apt-get update\nRUN apt-get install -y curl\nEOF\n && \
        docker build -t ${LOCAL_REGISTRY}/${project}/${repo}:${tag} . && \
        docker push ${LOCAL_REGISTRY}/${project}/${repo}:${tag}
    Execute Command In Host    ${cmd}
```

## Manual Test Verification

### Step 1: Create Test Projects
```bash
# Using Harbor CLI or UI
harbor project create test-dockerfile-with-label
harbor project create test-dockerfile-without-label
```

### Step 2: Build and Push Images
See build commands above for both image types.

### Step 3: Verify in Harbor UI
1. Navigate to test project → repository
2. Click on image to view artifact details
3. **For image WITH label:** Dockerfile tab shows content
4. **For image WITHOUT label:** Dockerfile tab shows info message + link to Build History

### Step 4: Verify via API
```bash
# Check if Dockerfile addition is advertised (image with label)
curl -u admin:Harbor12345 \
  https://harbor.example.com/api/v2.0/projects/test-dockerfile-with-label/repositories/test-image/artifacts/latest

# Should include: "addition_links": {"dockerfile": {...}}

# Fetch Dockerfile content
curl -u admin:Harbor12345 \
  https://harbor.example.com/api/v2.0/projects/test-dockerfile-with-label/repositories/test-image/artifacts/latest/additions/dockerfile
```

## Dockerfile Label Format

The Dockerfile content should be stored in one of these labels:
- `org.opencontainers.image.source` (recommended, OCI spec)
- `com.example.dockerfile` (custom)
- `dockerfile` (simple key)

Content can be:
- Plain Dockerfile text
- Base64 encoded (will be decoded by backend)
- Multi-line string (with `\n` separators)

Example with OCI-compliant label:
```dockerfile
LABEL org.opencontainers.image.source="\
FROM ubuntu:22.04\n\
RUN apt-get update && apt-get install -y curl\n\
COPY app.py /app/\n\
CMD [\"python3\", \"app.py\"]\
"
```

## Troubleshooting

### Dockerfile Tab Not Appearing
1. Verify image has label: `docker inspect <image> | grep -i dockerfile`
2. Check backend logs for parsing errors
3. Verify label key matches one of: `org.opencontainers.image.source`, `com.example.dockerfile`, `dockerfile`

### Content Not Displaying
1. Check browser console for JavaScript errors
2. Verify image label contains valid Dockerfile syntax
3. Check Content-Type header in API response (should be `text/plain; charset=utf-8`)

### Build History Fallback Not Working
1. Verify image has history (most images do)
2. Check build-history addition is advertised
3. Inspect image config: `docker inspect <image> | grep -A50 '"History"'`

## Performance Considerations

- Test images should be small (<100MB) for fast CI/CD
- Use minimal base image: `ubuntu:22.04` or `alpine:latest`
- Dockerfile content should be reasonable length (<10KB)

## Cleanup

Remove test images after test run:
```bash
docker rmi harbor.example.com/test-dockerfile-with-label/test-image:latest
docker rmi harbor.example.com/test-dockerfile-without-label/test-image:latest

# In Harbor UI, delete projects:
# - test-dockerfile-with-label
# - test-dockerfile-without-label
```
