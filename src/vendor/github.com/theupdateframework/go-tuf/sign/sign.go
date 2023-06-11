package sign

import (
	"encoding/json"

	"github.com/secure-systems-lab/go-securesystemslib/cjson"
	"github.com/theupdateframework/go-tuf/data"
	"github.com/theupdateframework/go-tuf/pkg/keys"
)

func Sign(s *data.Signed, k keys.Signer) error {
	ids := k.PublicData().IDs()
	signatures := make([]data.Signature, 0, len(s.Signatures)+1)
	for _, sig := range s.Signatures {
		found := false
		for _, id := range ids {
			if sig.KeyID == id {
				found = true
				break
			}
		}
		if !found {
			signatures = append(signatures, sig)
		}
	}

	canonical, err := cjson.EncodeCanonical(s.Signed)
	if err != nil {
		return err
	}

	sig, err := k.SignMessage(canonical)
	if err != nil {
		return err
	}

	s.Signatures = signatures
	for _, id := range ids {
		s.Signatures = append(s.Signatures, data.Signature{
			KeyID:     id,
			Signature: sig,
		})
	}

	return nil
}

func Marshal(v interface{}, keys ...keys.Signer) (*data.Signed, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	s := &data.Signed{Signed: b}
	for _, k := range keys {
		if err := Sign(s, k); err != nil {
			return nil, err
		}

	}
	return s, nil
}
