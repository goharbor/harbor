#!/bin/bash
set -e

REDHAT_PKGS="podman podman-docker podman-compose make"

DATA_VOLUME=${DATA_VOLUME:-/data}
if [ -d "$DATA_VOLUME" ]; then
    echo "Data volume $DATA_VOLUME exists. Attempting to stop any docker compose stack inside it."
    # Only try to run docker compose down if a compose file is present
    if [ -f "$DATA_VOLUME/docker-compose.yml" ] || [ -f "$DATA_VOLUME/docker-compose.yaml" ] || [ -f "$DATA_VOLUME/docker-compose.override.yml" ]; then
        if command -v docker >/dev/null 2>&1; then
            (cd "$DATA_VOLUME" && sudo docker compose down) || echo "docker compose down failed or no running compose"
        else
            echo "docker not found, skipping docker compose down"
        fi
    else
        echo "No docker compose files found in $DATA_VOLUME, skipping docker compose down"
    fi
    sudo rm -rf "$DATA_VOLUME"
fi
sudo mkdir -p "$DATA_VOLUME"
sudo mkdir -p /var/log/harbor
# Install required Red Hat packages only if any are missing
# Build a list of missing packages by checking with rpm -q
missing=""
for pkg in $REDHAT_PKGS; do
    if ! rpm -q "$pkg" &>/dev/null; then
        missing="$missing $pkg"
    fi
done
if [ -n "$(echo $missing | tr -d ' ')" ]; then
    echo "Installing missing packages:$missing"
    sudo dnf install -y $missing
else
    echo "All required packages are already installed: $REDHAT_PKGS"
fi
sudo sed -i 's/unqualified-search-registries = \["registry.access.redhat.com", "registry.redhat.io", "docker.io"\]/unqualified-search-registries = ["docker.io"]/g' /etc/containers/registries.conf

sudo cp ./make/common.sh ./make/harbor.yml.tmpl ./make/install.sh ./make/prepare "$DATA_VOLUME"/

# sudo chown -R 1000:1000 "$DATA_VOLUME"
sudo semanage fcontext -a -t container_file_t "$DATA_VOLUME(/.*)?"
sudo restorecon -R "$DATA_VOLUME"/
sudo make -f make/photon/Makefile _build_prepare -e BUILD_BASE=true BASEIMAGETAG=dev VERSIONTAG=dev
sudo mkdir -p "$DATA_VOLUME/common/config"
sudo mkdir -p "$DATA_VOLUME/cert"
# Build subjectAltName including loopback and all host IPv4 addresses
# Prefer `ip` to list addresses; fall back to `hostname -I` if needed.
ips=$(ip -o -4 addr show 2>/dev/null | awk '{print $4}' | cut -d/ -f1 | sort -u || true)
if [ -z "$ips" ]; then
    ips=$(hostname -I 2>/dev/null | tr ' ' '\n' | sed '/^$/d' | sort -u || true)
fi
san="subjectAltName=IP:127.0.0.1"
for ip in $ips; do
    # skip empty entries
    [ -z "$ip" ] && continue
    san="$san,IP:$ip"
done

sudo openssl req -newkey rsa:4096 -nodes -x509 -days 30 \
    -subj "/C=AU/ST=Victoria/L=Melbourne/O=deamen/CN=$(hostname)" \
    -addext "$san" \
    -keyout "$DATA_VOLUME/cert/harbor.key" \
    -out "$DATA_VOLUME/cert/harbor.crt"

sudo cp "$DATA_VOLUME/harbor.yml.tmpl" "$DATA_VOLUME/harbor.yml"
sudo sed -i "s|^hostname:.*|hostname: $(hostname)|" "$DATA_VOLUME/harbor.yml"
sudo sed -i "s|^  certificate:.*|  certificate: $DATA_VOLUME/cert/harbor.crt|" "$DATA_VOLUME/harbor.yml"
sudo sed -i "s|^  private_key:.*|  private_key: $DATA_VOLUME/cert/harbor.key|" "$DATA_VOLUME/harbor.yml"

if ! grep -qF 'localhost/goharbor/prepare:dev' "$DATA_VOLUME/prepare"; then
    sudo sed -i 's|goharbor/prepare:dev|localhost/goharbor/prepare:dev|g' "$DATA_VOLUME/prepare"
fi
cd "$DATA_VOLUME"

sudo ./install.sh --with-trivy --with-podman

# Push localhost/goharbor/prepare:dev to 127.0.0.1/library/prepare:dev
# Username: admin, password from $DATA_VOLUME/harbor.yml (harbor_admin_password)
HARBOR_YML="$DATA_VOLUME/harbor.yml"
if [ ! -f "$HARBOR_YML" ]; then
    echo "WARNING: $HARBOR_YML not found; cannot read harbor_admin_password. Skipping image push."
else
    harbor_admin_password=$(sed -n 's/^[[:space:]]*harbor_admin_password:[[:space:]]*//p' "$HARBOR_YML" | sed 's/^"\(.*\)"$/\1/;s/^'"'"'\(.*\)'"'"'$/\1/')
    harbor_admin_password=$(echo "$harbor_admin_password" | sed 's/^\s*//;s/\s*$//')
    if [ -z "$harbor_admin_password" ]; then
        echo "WARNING: harbor_admin_password is empty in $HARBOR_YML; skipping image push."
    else
        SRC_IMAGE="localhost/goharbor/prepare:dev"
        DST_IMAGE="127.0.0.1/library/prepare:dev"

        if ! sudo podman images --format '{{.Repository}}:{{.Tag}}' | grep -q "^$SRC_IMAGE$"; then
            echo "Source image $SRC_IMAGE not found locally; skipping push."
        else
            # Try to login and push with retries since Harbor services may take a moment to become ready
            max_retries=12
            delay=5
            pushed=0
            for i in $(seq 1 $max_retries); do
                echo "Attempt $i/$max_retries: pushing $SRC_IMAGE -> $DST_IMAGE"
                # disable xtrace so password isn't printed
                set +x
                if sudo podman login 127.0.0.1 -u admin -p "$harbor_admin_password" --tls-verify=false; then
                    sudo podman tag "$SRC_IMAGE" "$DST_IMAGE"
                    if sudo podman push --tls-verify=false "$DST_IMAGE"; then
                        echo "Image pushed successfully to $DST_IMAGE"
                        pushed=1
                        break
                    else
                        echo "podman push failed; will retry after $delay seconds"
                    fi
                else
                    echo "podman login failed; will retry after $delay seconds"
                fi
                sleep $delay
            done
            if [ "$pushed" -ne 1 ]; then
                echo "ERROR: failed to push $SRC_IMAGE to $DST_IMAGE after $max_retries attempts"
            fi
        fi
    fi
fi
