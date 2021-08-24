package export

import (
	"crypto/sha256"
	"github.com/opencontainers/go-digest"
	"io"
	"os"
)

// ArtifactDigestCalculator is an interface to be implemented by all file hash calculators
type ArtifactDigestCalculator interface {
	// Calculate returns the hash for a file
	Calculate(fileName string) (digest.Digest, error)
}

type SHA256ArtifactDigestCalculator struct{}

func (calc *SHA256ArtifactDigestCalculator) Calculate(fileName string) (digest.Digest, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return digest.NewDigest(digest.SHA256, hash), nil
}
