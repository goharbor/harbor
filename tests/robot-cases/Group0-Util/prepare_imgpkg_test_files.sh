#!/bin/bash
set -x
set -e

FILE_DIRECTORY=$1
FILE_PATH=$FILE_DIRECTORY/.imgpkg
mkdir "$FILE_DIRECTORY"
mkdir "$FILE_PATH"

cat > "$FILE_PATH"/bundle.yml <<EOF
---
apiVersion: imgpkg.carvel.dev/v1alpha1
kind: Bundle
metadata:
  name: my-app
authors:
- name: blah
  email: blah@blah.com
websites:
- url: blah.com
EOF

cat > "$FILE_PATH"/images.yml <<EOF
---
apiVersion: imgpkg.carvel.dev/v1alpha1
kind: ImagesLock
EOF