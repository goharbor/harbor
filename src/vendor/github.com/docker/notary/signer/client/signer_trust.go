// A CryptoService client wrapper around a remote wrapper service.

package client

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/docker/notary"
	pb "github.com/docker/notary/proto"
	"github.com/docker/notary/tuf/data"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// RemotePrivateKey is a key that is on a remote service, so no private
// key bytes are available
type RemotePrivateKey struct {
	data.PublicKey
	sClient pb.SignerClient
}

// RemoteSigner wraps a RemotePrivateKey and implements the crypto.Signer
// interface
type RemoteSigner struct {
	RemotePrivateKey
}

// Public method of a crypto.Signer needs to return a crypto public key.
func (rs *RemoteSigner) Public() crypto.PublicKey {
	publicKey, err := x509.ParsePKIXPublicKey(rs.RemotePrivateKey.Public())
	if err != nil {
		return nil
	}

	return publicKey
}

// NewRemotePrivateKey returns RemotePrivateKey, a data.PrivateKey that is only
// good for signing. (You can't get the private bytes out for instance.)
func NewRemotePrivateKey(pubKey data.PublicKey, sClient pb.SignerClient) *RemotePrivateKey {
	return &RemotePrivateKey{
		PublicKey: pubKey,
		sClient:   sClient,
	}
}

// Private returns nil bytes
func (pk *RemotePrivateKey) Private() []byte {
	return nil
}

// Sign calls a remote service to sign a message.
func (pk *RemotePrivateKey) Sign(rand io.Reader, msg []byte,
	opts crypto.SignerOpts) ([]byte, error) {

	keyID := pb.KeyID{ID: pk.ID()}
	sr := &pb.SignatureRequest{
		Content: msg,
		KeyID:   &keyID,
	}
	sig, err := pk.sClient.Sign(context.Background(), sr)
	if err != nil {
		return nil, err
	}
	return sig.Content, nil
}

// SignatureAlgorithm returns the signing algorithm based on the type of
// PublicKey algorithm.
func (pk *RemotePrivateKey) SignatureAlgorithm() data.SigAlgorithm {
	switch pk.PublicKey.Algorithm() {
	case data.ECDSAKey, data.ECDSAx509Key:
		return data.ECDSASignature
	case data.RSAKey, data.RSAx509Key:
		return data.RSAPSSSignature
	case data.ED25519Key:
		return data.EDDSASignature
	default: // unknown
		return ""
	}
}

// CryptoSigner returns a crypto.Signer tha wraps the RemotePrivateKey. Needed
// for implementing the interface.
func (pk *RemotePrivateKey) CryptoSigner() crypto.Signer {
	return &RemoteSigner{RemotePrivateKey: *pk}
}

// NotarySigner implements a RPC based Trust service that calls the Notary-signer Service
type NotarySigner struct {
	kmClient pb.KeyManagementClient
	sClient  pb.SignerClient

	healthClient healthpb.HealthClient
}

func healthCheck(d time.Duration, hc healthpb.HealthClient, serviceName string) (*healthpb.HealthCheckResponse, error) {
	ctx, _ := context.WithTimeout(context.Background(), d)
	req := &healthpb.HealthCheckRequest{
		Service: serviceName,
	}
	return hc.Check(ctx, req)
}

func healthCheckKeyManagement(d time.Duration, hc healthpb.HealthClient) error {
	out, err := healthCheck(d, hc, notary.HealthCheckKeyManagement)
	if err != nil {
		return err
	}
	if out.Status != healthpb.HealthCheckResponse_SERVING {
		return fmt.Errorf("Got the serving status of %s: %s, want %s", "KeyManagement", out.Status, healthpb.HealthCheckResponse_SERVING)
	}
	return nil
}

func healthCheckSigner(d time.Duration, hc healthpb.HealthClient) error {
	out, err := healthCheck(d, hc, notary.HealthCheckSigner)
	if err != nil {
		return err
	}
	if out.Status != healthpb.HealthCheckResponse_SERVING {
		return fmt.Errorf("Got the serving status of %s: %s, want %s", "Signer", out.Status, healthpb.HealthCheckResponse_SERVING)
	}
	return nil
}

// CheckHealth are used to probe whether the server is able to handle rpcs.
func (trust *NotarySigner) CheckHealth(d time.Duration, serviceName string) error {
	switch serviceName {
	case notary.HealthCheckKeyManagement:
		return healthCheckKeyManagement(d, trust.healthClient)
	case notary.HealthCheckSigner:
		return healthCheckSigner(d, trust.healthClient)
	case notary.HealthCheckOverall:
		if err := healthCheckKeyManagement(d, trust.healthClient); err != nil {
			return err
		}
		return healthCheckSigner(d, trust.healthClient)
	default:
		return fmt.Errorf("Unknown grpc service %s", serviceName)
	}
}

// NewGRPCConnection is a convenience method that returns GRPC Client Connection given a hostname, endpoint, and TLS options
func NewGRPCConnection(hostname string, port string, tlsConfig *tls.Config) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	netAddr := net.JoinHostPort(hostname, port)
	creds := credentials.NewTLS(tlsConfig)
	opts = append(opts, grpc.WithTransportCredentials(creds))
	return grpc.Dial(netAddr, opts...)
}

// NewNotarySigner is a convenience method that returns NotarySigner given a GRPC connection
func NewNotarySigner(conn *grpc.ClientConn) *NotarySigner {
	kmClient := pb.NewKeyManagementClient(conn)
	sClient := pb.NewSignerClient(conn)
	hc := healthpb.NewHealthClient(conn)

	return &NotarySigner{
		kmClient:     kmClient,
		sClient:      sClient,
		healthClient: hc,
	}
}

// Create creates a remote key and returns the PublicKey associated with the remote private key
func (trust *NotarySigner) Create(role data.RoleName, gun data.GUN, algorithm string) (data.PublicKey, error) {
	publicKey, err := trust.kmClient.CreateKey(context.Background(),
		&pb.CreateKeyRequest{Algorithm: algorithm, Role: role.String(), Gun: gun.String()})
	if err != nil {
		return nil, err
	}
	public := data.NewPublicKey(publicKey.KeyInfo.Algorithm.Algorithm, publicKey.PublicKey)
	return public, nil
}

// AddKey adds a key
func (trust *NotarySigner) AddKey(role data.RoleName, gun data.GUN, k data.PrivateKey) error {
	return errors.New("Adding a key to NotarySigner is not supported")
}

// RemoveKey deletes a key by ID - if the key didn't exist, succeed anyway
func (trust *NotarySigner) RemoveKey(keyid string) error {
	_, err := trust.kmClient.DeleteKey(context.Background(), &pb.KeyID{ID: keyid})
	return err
}

// GetKey retrieves a key by ID - returns nil if the key doesn't exist
func (trust *NotarySigner) GetKey(keyid string) data.PublicKey {
	pubKey, _, err := trust.getKeyInfo(keyid)
	if err != nil {
		return nil
	}
	return pubKey
}

func (trust *NotarySigner) getKeyInfo(keyid string) (data.PublicKey, data.RoleName, error) {
	keyInfo, err := trust.kmClient.GetKeyInfo(context.Background(), &pb.KeyID{ID: keyid})
	if err != nil {
		return nil, "", err
	}
	return data.NewPublicKey(keyInfo.KeyInfo.Algorithm.Algorithm, keyInfo.PublicKey), data.RoleName(keyInfo.Role), nil
}

// GetPrivateKey retrieves by ID an object that can be used to sign, but that does
// not contain any private bytes.  If the key doesn't exist, returns an error.
func (trust *NotarySigner) GetPrivateKey(keyid string) (data.PrivateKey, data.RoleName, error) {
	pubKey, role, err := trust.getKeyInfo(keyid)
	if err != nil {
		return nil, "", err
	}
	return NewRemotePrivateKey(pubKey, trust.sClient), role, nil
}

// ListKeys not supported for NotarySigner
func (trust *NotarySigner) ListKeys(role data.RoleName) []string {
	return []string{}
}

// ListAllKeys not supported for NotarySigner
func (trust *NotarySigner) ListAllKeys() map[string]data.RoleName {
	return map[string]data.RoleName{}
}
