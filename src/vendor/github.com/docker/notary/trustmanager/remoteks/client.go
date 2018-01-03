package remoteks

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/docker/notary/trustmanager"
)

// DefaultTimeout is the time a request will block waiting for a response
// from the server if no other timeout is configured.
const DefaultTimeout = time.Second * 30

// RemoteStore is a wrapper around the GRPC storage client, translating between
// the Go and GRPC APIs.
type RemoteStore struct {
	client   StoreClient
	location string
	timeout  time.Duration
}

var _ trustmanager.Storage = &RemoteStore{}

// NewRemoteStore instantiates a RemoteStore.
func NewRemoteStore(server string, tlsConfig *tls.Config, timeout time.Duration) (*RemoteStore, error) {
	cc, err := grpc.Dial(
		server,
		grpc.WithTransportCredentials(
			credentials.NewTLS(tlsConfig),
		),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	return &RemoteStore{
		client:   NewStoreClient(cc),
		location: server,
		timeout:  timeout,
	}, nil
}

// getContext returns a context with the timeout configured at initialization
// time of the RemoteStore. We currently throw away the cancel function as we
// have no current use cases to cancel sooner than the timeout.
func (s *RemoteStore) getContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), s.timeout)
	return ctx
}

// Set stores the data using the provided fileName
func (s *RemoteStore) Set(fileName string, data []byte) error {
	sm := &SetMsg{
		FileName: fileName,
		Data:     data,
	}
	_, err := s.client.Set(s.getContext(), sm)
	return err
}

// Remove deletes a file from the store relative to the store's base directory.
// Paths are expected to be cleaned server side.
func (s *RemoteStore) Remove(fileName string) error {
	fm := &FileNameMsg{
		FileName: fileName,
	}
	_, err := s.client.Remove(s.getContext(), fm)
	return err
}

// Get returns the file content found at fileName relative to the base directory
// of the file store. Paths are expected to be cleaned server side.
func (s *RemoteStore) Get(fileName string) ([]byte, error) {
	fm := &FileNameMsg{
		FileName: fileName,
	}
	bm, err := s.client.Get(s.getContext(), fm)
	if err != nil {
		return nil, err
	}
	return bm.Data, nil
}

// ListFiles returns a list of paths relative to the base directory of the
// filestore. Any of these paths must be retrievable via the
// Storage.Get method.
func (s *RemoteStore) ListFiles() []string {
	logrus.Infof("listing files from %s", s.location)
	fl, err := s.client.ListFiles(s.getContext(), &google_protobuf.Empty{})
	if err != nil {
		logrus.Errorf("error listing files from %s: %s", s.location, err.Error())
		return nil
	}
	return fl.FileNames

}

// Location returns a human readable indication of where the storage is located.
func (s *RemoteStore) Location() string {
	return fmt.Sprintf("Remote Key Store @ %s", s.location)
}
