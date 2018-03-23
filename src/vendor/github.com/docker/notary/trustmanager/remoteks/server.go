package remoteks

import (
	"github.com/Sirupsen/logrus"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"

	"github.com/docker/notary/trustmanager"
)

// GRPCStorage is an implementer of the GRPC storage server. It passes through
// the requested operations to an underlying trustmanager.Storage instance, translating
// between the Go and GRPC interfaces.
type GRPCStorage struct {
	backend trustmanager.Storage
}

// NewGRPCStorage instantiates a new GRPC storage server using the provided
// backend.
func NewGRPCStorage(backend trustmanager.Storage) *GRPCStorage {
	return &GRPCStorage{
		backend: backend,
	}
}

// Set writes the provided data under the given identifier.
func (s *GRPCStorage) Set(ctx context.Context, msg *SetMsg) (*google_protobuf.Empty, error) {
	logrus.Debugf("storing: %s", msg.FileName)
	err := s.backend.Set(msg.FileName, msg.Data)
	if err != nil {
		logrus.Errorf("failed to store: %s", err.Error())
	}
	return &google_protobuf.Empty{}, err
}

// Remove deletes the data associated with the provided identifier.
func (s *GRPCStorage) Remove(ctx context.Context, fn *FileNameMsg) (*google_protobuf.Empty, error) {
	return &google_protobuf.Empty{}, s.backend.Remove(fn.FileName)
}

// Get returns the data associated with the provided identifier.
func (s *GRPCStorage) Get(ctx context.Context, fn *FileNameMsg) (*ByteMsg, error) {
	data, err := s.backend.Get(fn.FileName)
	if err != nil {
		return &ByteMsg{}, err
	}
	return &ByteMsg{Data: data}, nil
}

// ListFiles returns all known identifiers in the storage backend.
func (s *GRPCStorage) ListFiles(ctx context.Context, _ *google_protobuf.Empty) (*StringListMsg, error) {
	lst := s.backend.ListFiles()
	logrus.Debugf("found %d keys", len(lst))
	return &StringListMsg{
		FileNames: lst,
	}, nil
}
