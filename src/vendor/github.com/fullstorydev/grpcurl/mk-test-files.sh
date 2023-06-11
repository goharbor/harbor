#!/bin/bash

set -e

cd "$(dirname $0)"

# Run this script to generate files used by tests.

echo "Creating protosets..."
protoc testing/test.proto \
	--include_imports \
	--descriptor_set_out=testing/test.protoset

protoc testing/example.proto \
	--include_imports \
	--descriptor_set_out=testing/example.protoset

protoc testing/jsonpb_test_proto/test_objects.proto \
	--go_out=paths=source_relative:.

echo "Creating certs for TLS testing..."
if ! hash certstrap 2>/dev/null; then
  # certstrap not found: try to install it
  go get github.com/square/certstrap
  go install github.com/square/certstrap
fi

function cs() {
	certstrap --depot-path testing/tls "$@" --passphrase ""
}

rm -rf testing/tls

# Create CA
cs init --years 10 --common-name ca

# Create client cert
cs request-cert --common-name client
cs sign client --years 10 --CA ca

# Create server cert
cs request-cert --common-name server --ip 127.0.0.1 --domain localhost
cs sign server --years 10 --CA ca

# Create another server cert for error testing
cs request-cert --common-name other --ip 1.2.3.4 --domain foobar.com
cs sign other --years 10 --CA ca

# Create another CA and client cert for more
# error testing
cs init --years 10 --common-name wrong-ca
cs request-cert --common-name wrong-client
cs sign wrong-client --years 10 --CA wrong-ca

# Create expired cert
cs request-cert --common-name expired --ip 127.0.0.1 --domain localhost
cs sign expired --years 0 --CA ca
