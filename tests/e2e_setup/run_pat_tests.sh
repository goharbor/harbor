#!/bin/bash
# Run PAT robot tests without Docker daemon complications

set -e

HARBOR_SRC_FOLDER=$(cd "$(dirname "$0")/../../" && pwd)
TEST_SETUP_DIR=$(cd "$(dirname "$0")" && pwd)

echo "=============================================="
echo "Harbor PAT Robot Tests"
echo "=============================================="
echo ""
echo "Harbor source: ${HARBOR_SRC_FOLDER}"
echo "Test setup: ${TEST_SETUP_DIR}"
echo ""

# Mount Docker socket if available
DOCKER_MOUNT=""
if [ -S /var/run/docker.sock ]; then
    DOCKER_MOUNT="-v /var/run/docker.sock:/var/run/docker.sock:rw"
    echo "✓ Docker socket available"
else
    echo "⚠ Docker socket not available (some tests may be skipped)"
fi

echo ""
echo "Starting E2E container and running PAT tests..."
echo ""

# Run the container with proper setup
docker run --rm --privileged \
  -v /var/log/harbor:/var/log/harbor \
  -v /etc/hosts:/etc/hosts \
  -v ${HARBOR_SRC_FOLDER}:/drone \
  -v ${HARBOR_SRC_FOLDER}/tests/harbor_ca.crt:/ca/ca.crt \
  -v /dev/shm:/dev/shm \
  ${DOCKER_MOUNT} \
  -e NETWORK_TYPE=public \
  -e HARBOR_ADMIN=admin \
  -e HARBOR_PASSWORD=Harbor12345 \
  -w /drone \
  registry.goharbor.io/harbor-ci/goharbor/harbor-e2e-engine:latest-ui \
  bash -c "
    echo 'Container started'
    echo 'Running PAT tests...'
    echo ''

    # Run robot tests for PAT functionality
    robot \
      -V /drone/tests/e2e_setup/robotvars.py \
      -d /drone/test-results \
      /drone/tests/robot-cases/Group1-Nightly/PAT.robot
  "

echo ""
echo "=============================================="
echo "Test Results:"
echo "=============================================="
echo ""
echo "Test results saved to: /home/rossg/src/harbor/test-results/"
echo ""
if [ -f "${HARBOR_SRC_FOLDER}/test-results/log.html" ]; then
    echo "✓ Log file: test-results/log.html"
    echo "✓ Report: test-results/report.html"
else
    echo "Check the output above for test results"
fi
