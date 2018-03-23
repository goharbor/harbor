package keydbstore

import (
	"sync"

	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
)

type cachedKeyService struct {
	signed.CryptoService
	lock       *sync.Mutex
	cachedKeys map[string]*cachedKey
}

type cachedKey struct {
	role data.RoleName
	key  data.PrivateKey
}

// NewCachedKeyService returns a new signed.CryptoService that includes caching
func NewCachedKeyService(baseKeyService signed.CryptoService) signed.CryptoService {
	return &cachedKeyService{
		CryptoService: baseKeyService,
		lock:          &sync.Mutex{},
		cachedKeys:    make(map[string]*cachedKey),
	}
}

// AddKey stores the contents of a private key. Both role and gun are ignored,
// we always use Key IDs as name, and don't support aliases
func (s *cachedKeyService) AddKey(role data.RoleName, gun data.GUN, privKey data.PrivateKey) error {
	if err := s.CryptoService.AddKey(role, gun, privKey); err != nil {
		return err
	}

	// Add the private key to our cache
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cachedKeys[privKey.ID()] = &cachedKey{
		role: role,
		key:  privKey,
	}

	return nil
}

// GetKey returns the PrivateKey given a KeyID
func (s *cachedKeyService) GetPrivateKey(keyID string) (data.PrivateKey, data.RoleName, error) {
	cachedKeyEntry, ok := s.cachedKeys[keyID]
	if ok {
		return cachedKeyEntry.key, cachedKeyEntry.role, nil
	}

	// retrieve the key from the underlying store and put it into the cache
	privKey, role, err := s.CryptoService.GetPrivateKey(keyID)
	if err == nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		// Add the key to cache
		s.cachedKeys[privKey.ID()] = &cachedKey{key: privKey, role: role}
		return privKey, role, nil
	}
	return nil, "", err
}

// RemoveKey removes the key from the keyfilestore
func (s *cachedKeyService) RemoveKey(keyID string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.cachedKeys, keyID)
	return s.CryptoService.RemoveKey(keyID)
}
