#!/bin/bash
# Copyright Project Harbor Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

# Configuration
HARBOR_REGISTRY="${HARBOR_REGISTRY:-localhost:5000}"
DOCKER_BUILDKIT="${DOCKER_BUILDKIT:-1}"
PUSH_IMAGES="${PUSH_IMAGES:-true}"
CLEANUP_AFTER="${CLEANUP_AFTER:-false}"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Create test image with Dockerfile label
build_image_with_dockerfile_label() {
    local project="test-dockerfile-with-label"
    local repo="test-image"
    local tag="latest"
    local workdir="/tmp/harbor-test-dockerfile-with-label"

    log_info "Building test image WITH Dockerfile label..."
    log_info "  Project: $project"
    log_info "  Repository: $repo"
    log_info "  Tag: $tag"

    # Create work directory
    mkdir -p "$workdir"
    cd "$workdir"

    # Create Dockerfile
    cat > Dockerfile << 'EOF'
FROM ubuntu:22.04
RUN apt-get update
RUN apt-get install -y curl
COPY test.txt /app/
LABEL org.opencontainers.image.authors="Harbor Test Suite"
LABEL org.opencontainers.image.description="Test image for Dockerfile display feature"
EOF

    # Create test file
    echo "Test content for artifact display" > test.txt

    # Read Dockerfile content for label
    dockerfile_content=$(cat Dockerfile)

    # Build image with Dockerfile label
    DOCKER_BUILDKIT=$DOCKER_BUILDKIT docker build \
        --label "org.opencontainers.image.source=${dockerfile_content}" \
        -t "${HARBOR_REGISTRY}/${project}/${repo}:${tag}" \
        .

    log_info "Image built successfully: ${HARBOR_REGISTRY}/${project}/${repo}:${tag}"

    # Verify label exists
    log_info "Verifying Dockerfile label..."
    docker inspect "${HARBOR_REGISTRY}/${project}/${repo}:${tag}" | grep -A5 "org.opencontainers.image.source" || log_warn "Dockerfile label may not be visible in inspect (this is OK)"

    # Push if requested
    if [ "$PUSH_IMAGES" == "true" ]; then
        log_info "Pushing image to registry..."
        docker push "${HARBOR_REGISTRY}/${project}/${repo}:${tag}"
        log_info "Image pushed successfully"
    fi

    cd - > /dev/null
}

# Create test image without Dockerfile label
build_image_without_dockerfile_label() {
    local project="test-dockerfile-without-label"
    local repo="test-image"
    local tag="latest"
    local workdir="/tmp/harbor-test-dockerfile-without-label"

    log_info "Building test image WITHOUT Dockerfile label..."
    log_info "  Project: $project"
    log_info "  Repository: $repo"
    log_info "  Tag: $tag"

    # Create work directory
    mkdir -p "$workdir"
    cd "$workdir"

    # Create Dockerfile (same content as above, but no Dockerfile label)
    cat > Dockerfile << 'EOF'
FROM ubuntu:22.04
RUN apt-get update
RUN apt-get install -y curl
COPY test.txt /app/
LABEL org.opencontainers.image.authors="Harbor Test Suite"
LABEL org.opencontainers.image.description="Test image without Dockerfile label"
EOF

    # Create test file
    echo "Test content for artifact display" > test.txt

    # Build image WITHOUT Dockerfile label
    DOCKER_BUILDKIT=$DOCKER_BUILDKIT docker build \
        -t "${HARBOR_REGISTRY}/${project}/${repo}:${tag}" \
        .

    log_info "Image built successfully: ${HARBOR_REGISTRY}/${project}/${repo}:${tag}"

    # Verify no Dockerfile label
    log_info "Verifying NO Dockerfile label..."
    if docker inspect "${HARBOR_REGISTRY}/${project}/${repo}:${tag}" | grep -i "org.opencontainers.image.source" > /dev/null; then
        log_warn "Unexpected: image has Dockerfile label when it shouldn't"
    else
        log_info "Confirmed: no Dockerfile label present"
    fi

    # Push if requested
    if [ "$PUSH_IMAGES" == "true" ]; then
        log_info "Pushing image to registry..."
        docker push "${HARBOR_REGISTRY}/${project}/${repo}:${tag}"
        log_info "Image pushed successfully"
    fi

    cd - > /dev/null
}

# Cleanup test images
cleanup_images() {
    log_info "Cleaning up test images..."

    docker rmi -f "${HARBOR_REGISTRY}/test-dockerfile-with-label/test-image:latest" || true
    docker rmi -f "${HARBOR_REGISTRY}/test-dockerfile-without-label/test-image:latest" || true

    log_info "Cleanup complete"
}

# Main execution
main() {
    log_info "Harbor Artifact Detail Additions - Test Image Builder"
    log_info "Registry: $HARBOR_REGISTRY"
    log_info "BuildKit: $DOCKER_BUILDKIT"
    log_info "Push images: $PUSH_IMAGES"
    log_info "Cleanup after: $CLEANUP_AFTER"
    echo ""

    # Check if Docker is available
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi

    # Build images
    build_image_with_dockerfile_label
    echo ""
    build_image_without_dockerfile_label
    echo ""

    # Cleanup if requested
    if [ "$CLEANUP_AFTER" == "true" ]; then
        cleanup_images
    fi

    log_info "Test image setup complete!"
    log_info ""
    log_info "Next steps:"
    log_info "1. Ensure Harbor projects exist:"
    log_info "   - test-dockerfile-with-label"
    log_info "   - test-dockerfile-without-label"
    log_info "2. Run E2E tests:"
    log_info "   robot tests/robot-cases/Group1-Nightly/Artifact-Detail-Additions.robot"
}

# Argument parsing
while [[ $# -gt 0 ]]; do
    case $1 in
        --registry)
            HARBOR_REGISTRY="$2"
            shift 2
            ;;
        --no-push)
            PUSH_IMAGES="false"
            shift
            ;;
        --cleanup)
            CLEANUP_AFTER="true"
            shift
            ;;
        --no-buildkit)
            DOCKER_BUILDKIT="0"
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --registry REGISTRY      Harbor registry (default: localhost:5000)"
            echo "  --no-push                Don't push images to registry"
            echo "  --cleanup                Remove test images after build"
            echo "  --no-buildkit            Disable Docker BuildKit"
            echo "  --help                   Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main
main
