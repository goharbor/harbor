package signer

import (
	"crypto/tls"

	pb "github.com/docker/notary/proto"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
)

// SigningService is the interface to implement a key management and signing service
type SigningService interface {
	KeyManager

	// Signer returns a Signer for a given keyID
	Signer(keyID *pb.KeyID) (Signer, error)
}

// CryptoServiceIndex represents a mapping between a service algorithm string
// and a CryptoService
type CryptoServiceIndex map[string]signed.CryptoService

// KeyManager is the interface to implement key management (possibly a key database)
type KeyManager interface {
	// CreateKey creates a new key and returns it's Information
	CreateKey() (*pb.PublicKey, error)

	// DeleteKey removes a key
	DeleteKey(keyID *pb.KeyID) (*pb.Void, error)

	// KeyInfo returns the public key of a particular key
	KeyInfo(keyID *pb.KeyID) (*pb.PublicKey, error)
}

// Signer is the interface that allows the signing service to return signatures
type Signer interface {
	Sign(request *pb.SignatureRequest) (*pb.Signature, error)
}

// Config tells how to configure a notary-signer
type Config struct {
	GRPCAddr       string
	TLSConfig      *tls.Config
	CryptoServices CryptoServiceIndex
	PendingKeyFunc func(trustmanager.KeyInfo) (data.PublicKey, error)
}
