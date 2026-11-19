#!/bin/bash

set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
test_root="$(mktemp -d)"
trap 'rm -rf "$test_root"' EXIT

cp "$repository_root/make/common.sh" "$test_root/common.sh"
cp "$repository_root/make/install.sh" "$test_root/install.sh"
cp "$repository_root/make/prepare" "$test_root/prepare"
cp "$repository_root/make/harbor.yml.tmpl" "$test_root/harbor.yml"

docker() {
    case "$*" in
        --version)
            echo "Docker version 28.0.0, build test"
            ;;
        "compose version")
            echo "Docker Compose version v2.39.0"
            ;;
        "compose ps -q"|"compose up -d")
            ;;
        run*)
            case " $* " in
                *" -e CONTAINER_RUNTIME=docker "*)
                    ;;
                *)
                    echo "Docker prepare did not receive its runtime selection." >&2
                    return 1
                    ;;
            esac
            case " $* " in
                *" --security-opt label=disable "*)
                    echo "Docker prepare unexpectedly disabled SELinux separation." >&2
                    return 1
                    ;;
                *" --privileged goharbor/prepare:dev prepare "*)
                    ;;
                *)
                    echo "Unexpected Docker prepare invocation: $*" >&2
                    return 1
                    ;;
            esac
            ;;
        *)
            echo "Unexpected Docker invocation: $*" >&2
            return 1
            ;;
    esac
}

podman() {
    case "${1:-}" in
        --version)
            echo "podman version 5.8.2"
            ;;
        run)
            case " $* " in
                *" --security-opt label=disable -e CONTAINER_RUNTIME=podman "*" --privileged goharbor/prepare:dev prepare "*)
                    ;;
                *)
                    echo "Unexpected prepare invocation: $*" >&2
                    return 1
                    ;;
            esac
            if [[ ${FIXTURE_PREPARE_FAILURE:-false} == true ]]; then return 42; fi
            ;;
        *)
            echo "Unexpected Podman invocation: $*" >&2
            return 1
            ;;
    esac
}

podman-compose() {
    case "$*" in
        --version)
            echo "podman-compose version 1.5.0"
            ;;
        "ps -q"|"up -d")
            ;;
        *)
            echo "Unexpected Podman Compose invocation: $*" >&2
            return 1
            ;;
    esac
}

export -f docker podman podman-compose
export TERM="${TERM:-xterm}"

CONTAINER_RUNTIME=podman "$test_root/install.sh"
CONTAINER_RUNTIME=docker "$test_root/install.sh"

# Execute the real installer/prepare control flow. A failed runtime preparation
# must propagate instead of starting Compose and printing installation success.
if CONTAINER_RUNTIME=podman FIXTURE_PREPARE_FAILURE=true "$test_root/install.sh" > "$test_root/failure.log" 2>&1; then
    echo "Failed Podman preparation was accepted by the installer." >&2
    exit 1
else
    prepare_status=$?
fi
if [[ $prepare_status -ne 42 ]]; then
    echo "Installer did not preserve the Podman preparation exit status." >&2
    exit 1
fi
if grep -q 'starting Harbor\|installed and started successfully' "$test_root/failure.log"; then
    echo "Installer continued after failed Podman preparation." >&2
    exit 1
fi
rm -rf "$test_root/input"

if CONTAINER_RUNTIME=unsupported "$test_root/install.sh" > "$test_root/unsupported.log" 2>&1; then
    echo "Unsupported runtime was accepted." >&2
    exit 1
fi
grep -q 'Unsupported container runtime' "$test_root/unsupported.log"
if [[ -e "$test_root/input" ]]; then
    echo "Unsupported runtime reached preparation." >&2
    exit 1
fi

python3 - "$repository_root/make/photon/prepare/templates/docker_compose/docker-compose.yml.jinja" <<'PY'
import pathlib
import sys

template = pathlib.Path(sys.argv[1]).read_text().splitlines()
bind_mounts = [index for index, line in enumerate(template) if line.strip() == "- type: bind"]
for index in bind_mounts:
    block = "\n".join(template[index:index + 6])
    if "bind:\n" not in block or "selinux: z" not in block:
        raise SystemExit(f"Bind mount at line {index + 1} lacks SELinux relabeling")
PY

if [ -e "$test_root/input" ]; then
    echo "The prepare input directory was not cleaned up." >&2
    exit 1
fi

echo "Docker and Podman installer runtime tests passed."
