package proto

// this file exists solely to allow us to use `go generate` to build our
// compiled GRPC interface for Go.
//go:generate protoc -I ./ ./signer.proto --go_out=plugins=grpc:.
