#!/usr/bin/env bash

# This script cross-compiles static (when possible) binaries for supported OS's
# architectures.  The Linux binary is completely static, whereas Mac OS binary
# has libtool statically linked in. but is otherwise not static because you
# cannot statically link to system libraries in Mac OS.

GOARCH="amd64"

for os in "$@"; do
	export GOOS="${os}"
	BUILDTAGS="${NOTARY_BUILDTAGS}"
	OUTFILE=notary

	if [[ "${GOOS}" == "darwin" ]]; then
		export CC="o64-clang"
		export CXX="o64-clang++"
		# darwin binaries can't be compiled to be completely static with the -static flag
		LDFLAGS=""
	else
		# no building with Cgo.  Also no building with pkcs11
		if [[ "${GOOS}" == "windows" ]]; then
			BUILDTAGS=
			OUTFILE=notary.exe
		fi
		unset CC
		unset CXX
		LDFLAGS="-extldflags -static"
	fi

	if [[ "${BUILDTAGS}" == *pkcs11* ]]; then
		export CGO_ENABLED=1
	else
		export CGO_ENABLED=0
	fi

	mkdir -p "${NOTARYDIR}/cross/${GOOS}/${GOARCH}";

	set -x;
	go build \
		-o "${NOTARYDIR}/cross/${GOOS}/${GOARCH}/${OUTFILE}" \
		-a \
		-tags "${BUILDTAGS} netgo" \
		-ldflags "-w ${CTIMEVAR} ${LDFLAGS}"  \
		./cmd/notary;
	set +x;
done
