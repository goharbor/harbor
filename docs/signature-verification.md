# Harbor Release Signature Verification

> **Note:** Signature verification is available starting with Harbor v2.15.0. Earlier releases are not signed.

## Table of Contents
- [Overview](#overview)
- [Why Verify](#why-verify)
- [Prerequisites](#prerequisites)
- [Verification Steps](#verification-steps)
- [Troubleshooting](#troubleshooting)
- [What Gets Verified](#what-gets-verified)
- [Resources](#resources)

## Overview

Harbor release artifacts (installers) are cryptographically signed using [Cosign](https://docs.sigstore.dev/cosign/overview/) with keyless signing. This allows you to verify that downloads are authentic and unmodified.

## Why Verify

* Confirms the file came from Harbor's official build  
* Detects any modifications or tampering  
* Protects against malicious downloads

## Prerequisites

**Install Cosign (v2.0+):**
```bash
# macOS
brew install sigstore/tap/cosign

# Linux
curl -LO https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64
chmod +x cosign-linux-amd64
sudo mv cosign-linux-amd64 /usr/local/bin/cosign

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/sigstore/cosign/releases/latest/download/cosign-windows-amd64.exe" -OutFile "cosign.exe"

# Verify installation
cosign version
```

## Verification Steps

### 1. Download Files
```bash
# Download installer and Signature file (example v2.15.0)
wget https://github.com/goharbor/harbor/releases/download/v2.15.0/harbor-offline-installer-v2.15.0.tgz
wget https://github.com/goharbor/harbor/releases/download/v2.15.0/harbor-offline-installer-v2.15.0.tgz.sigstore.json
```

### 2. Verify Signature
```bash
cosign verify-blob \
  --bundle harbor-offline-installer-v2.15.0.tgz.sigstore.json \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp '^https://github.com/goharbor/harbor/.github/workflows/publish_release.yml@refs/tags/v.*$' \
  harbor-offline-installer-v2.15.0.tgz
```

**Expected output:**
```
Verified OK
```

### 3. For Online Installer
```bash
wget https://github.com/goharbor/harbor/releases/download/v2.15.0/harbor-online-installer-v2.15.0.tgz
wget https://github.com/goharbor/harbor/releases/download/v2.15.0/harbor-online-installer-v2.15.0.tgz.sigstore.json

cosign verify-blob \
  --bundle harbor-online-installer-v2.15.0.tgz.sigstore.json \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity-regexp '^https://github.com/goharbor/harbor/.github/workflows/publish_release.yml@refs/tags/v.*$' \
  harbor-online-installer-v2.15.0.tgz
```

## Troubleshooting

### Certificate identity doesn't match
**Cause:** Incorrect repository name in verification command  
**Solution:** Ensure you're using `goharbor/harbor` in the `--certificate-identity-regexp` parameter

### Unable to find signature
**Cause:** Signature file not in the same directory as the installer  
**Solution:** Ensure both `.tgz` and `.tgz.sigstore.json` files are in the current working directory

### Bad signature
**Cause:** Downloaded files are corrupted or incomplete  
**Solution:** Re-download both the installer and signature files from the official [Harbor releases page](https://github.com/goharbor/harbor/releases)

### Version not supported
**Cause:** Attempting to verify releases prior to v2.15.0  
**Solution:** Signature verification is only available for Harbor v2.15.0 and later

## What Gets Verified

- **File authenticity** - Signed by official Harbor CI/CD workflow  
- **File integrity** - No modifications since signing  
- **Build provenance** - Logged in public Sigstore transparency log

## Resources

- [Cosign Documentation](https://docs.sigstore.dev/)
- [Harbor Releases](https://github.com/goharbor/harbor/releases)

---

**Applies to:** Harbor v2.15.0 and later
