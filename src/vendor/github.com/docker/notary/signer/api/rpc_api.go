package api

import (
	"crypto/rand"
	"fmt"

	ctxu "github.com/docker/distribution/context"
	"github.com/docker/notary/signer"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/docker/notary/proto"
)

//KeyManagementServer implements the KeyManagementServer grpc interface
type KeyManagementServer struct {
	CryptoServices signer.CryptoServiceIndex
}

//SignerServer implements the SignerServer grpc interface
type SignerServer struct {
	CryptoServices signer.CryptoServiceIndex
}

//CreateKey returns a PublicKey created using KeyManagementServer's SigningService
func (s *KeyManagementServer) CreateKey(ctx context.Context, req *pb.CreateKeyRequest) (*pb.PublicKey, error) {
	service := s.CryptoServices[req.Algorithm]

	logger := ctxu.GetLogger(ctx)

	if service == nil {
		logger.Error("CreateKey: unsupported algorithm: ", req.Algorithm)
		return nil, fmt.Errorf("algorithm %s not supported for create key", req.Algorithm)
	}

	var tufKey data.PublicKey
	var err error

	tufKey, err = service.Create(data.RoleName(req.Role), data.GUN(req.Gun), req.Algorithm)
	if err != nil {
		logger.Error("CreateKey: failed to create key: ", err)
		return nil, grpc.Errorf(codes.Internal, "Key creation failed")
	}
	logger.Info("CreateKey: Created KeyID ", tufKey.ID())

	return &pb.PublicKey{
		KeyInfo: &pb.KeyInfo{
			KeyID:     &pb.KeyID{ID: tufKey.ID()},
			Algorithm: &pb.Algorithm{Algorithm: tufKey.Algorithm()},
		},
		PublicKey: tufKey.Public(),
	}, nil
}

//DeleteKey deletes they key associated with a KeyID
func (s *KeyManagementServer) DeleteKey(ctx context.Context, keyID *pb.KeyID) (*pb.Void, error) {
	logger := ctxu.GetLogger(ctx)
	// delete key ID from all services
	for _, service := range s.CryptoServices {
		if err := service.RemoveKey(keyID.ID); err != nil {
			logger.Errorf("Failed to delete key %s", keyID.ID)
			return nil, grpc.Errorf(codes.Internal, "Key deletion for KeyID %s failed", keyID.ID)
		}
	}

	return &pb.Void{}, nil
}

//GetKeyInfo returns they PublicKey associated with a KeyID
func (s *KeyManagementServer) GetKeyInfo(ctx context.Context, keyID *pb.KeyID) (*pb.GetKeyInfoResponse, error) {
	privKey, role, err := findKeyByID(s.CryptoServices, keyID)

	logger := ctxu.GetLogger(ctx)

	if err != nil {
		logger.Errorf("GetKeyInfo: key %s not found", keyID.ID)
		return nil, grpc.Errorf(codes.NotFound, "key %s not found", keyID.ID)
	}

	logger.Debug("GetKeyInfo: Returning PublicKey for KeyID ", keyID.ID)
	return &pb.GetKeyInfoResponse{
		KeyInfo: &pb.KeyInfo{
			KeyID:     &pb.KeyID{ID: privKey.ID()},
			Algorithm: &pb.Algorithm{Algorithm: privKey.Algorithm()},
		},
		PublicKey: privKey.Public(),
		Role:      role.String(),
	}, nil
}

//Sign signs a message and returns the signature using a private key associate with the KeyID from the SignatureRequest
func (s *SignerServer) Sign(ctx context.Context, sr *pb.SignatureRequest) (*pb.Signature, error) {
	privKey, _, err := findKeyByID(s.CryptoServices, sr.KeyID)

	logger := ctxu.GetLogger(ctx)

	switch err.(type) {
	case trustmanager.ErrKeyNotFound:
		logger.Errorf("Sign: key %s not found", sr.KeyID.ID)
		return nil, grpc.Errorf(codes.NotFound, err.Error())
	case nil:
		break
	default:
		logger.Errorf("Getting key %s failed: %s", sr.KeyID.ID, err.Error())
		return nil, grpc.Errorf(codes.Internal, err.Error())

	}

	sig, err := privKey.Sign(rand.Reader, sr.Content, nil)
	if err != nil {
		logger.Errorf("Sign: signing failed for KeyID %s on hash %s", sr.KeyID.ID, sr.Content)
		return nil, grpc.Errorf(codes.Internal, "Signing failed for KeyID %s on hash %s", sr.KeyID.ID, sr.Content)
	}

	logger.Info("Sign: Signed ", string(sr.Content), " with KeyID ", sr.KeyID.ID)

	signature := &pb.Signature{
		KeyInfo: &pb.KeyInfo{
			KeyID:     &pb.KeyID{ID: privKey.ID()},
			Algorithm: &pb.Algorithm{Algorithm: privKey.Algorithm()},
		},
		Algorithm: &pb.Algorithm{Algorithm: privKey.SignatureAlgorithm().String()},
		Content:   sig,
	}

	return signature, nil
}
