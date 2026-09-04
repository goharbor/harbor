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
                *" --security-opt label=disable "*" --privileged goharbor/prepare:dev prepare "*)
                    ;;
                *)
                    echo "Unexpected prepare invocation: $*" >&2
                    return 1
                    ;;
            esac
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
