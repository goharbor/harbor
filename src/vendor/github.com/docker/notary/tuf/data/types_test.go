package data

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/docker/go/canonical/json"
	"github.com/docker/notary"
	"github.com/stretchr/testify/require"
)

func TestGenerateFileMetaDefault(t *testing.T) {
	// default is sha512
	r := bytes.NewReader([]byte("foo"))
	meta, err := NewFileMeta(r, notary.SHA512)
	require.NoError(t, err, "Unexpected error.")
	require.Equal(t, meta.Length, int64(3), "Meta did not have expected Length field value")
	hashes := meta.Hashes
	require.Len(t, hashes, 1, "Only expected one hash to be present")
	hash, ok := hashes[notary.SHA512]
	if !ok {
		t.Fatal("missing sha512 hash")
	}
	require.Equal(t, "f7fbba6e0636f890e56fbbf3283e524c6fa3204ae298382d624741d0dc6638326e282c41be5e4254d8820772c5518a2c5a8c0c7f7eda19594a7eb539453e1ed7", hex.EncodeToString(hash), "Hashes not equal")
}

func TestGenerateFileMetaExplicit(t *testing.T) {
	r := bytes.NewReader([]byte("foo"))
	meta, err := NewFileMeta(r, notary.SHA256, notary.SHA512)
	require.NoError(t, err)
	require.Equal(t, meta.Length, int64(3))
	hashes := meta.Hashes
	require.Len(t, hashes, 2)
	for name, val := range map[string]string{
		notary.SHA256: "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
		notary.SHA512: "f7fbba6e0636f890e56fbbf3283e524c6fa3204ae298382d624741d0dc6638326e282c41be5e4254d8820772c5518a2c5a8c0c7f7eda19594a7eb539453e1ed7",
	} {
		hash, ok := hashes[name]
		if !ok {
			t.Fatalf("missing %s hash", name)
		}
		require.Equal(t, hex.EncodeToString(hash), val)
	}
}

func TestSignatureUnmarshalJSON(t *testing.T) {
	signatureJSON := `{"keyid":"97e8e1b51b6e7cf8720a56b5334bd8692ac5b28233c590b89fab0b0cd93eeedc","method":"RSA","sig":"2230cba525e4f5f8fc744f234221ca9a92924da4cc5faf69a778848882fcf7a20dbb57296add87f600891f2569a9c36706314c240f9361c60fd36f5a915a0e9712fc437b761e8f480868d7a4444724daa0d29a2669c0edbd4046046649a506b3d711d0aa5e70cb9d09dec7381e7de27a3168e77731e08f6ed56fcce2478855e837816fb69aff53412477748cd198dce783850080d37aeb929ad0f81460ebd31e61b772b6c7aa56977c787d4281fa45dbdefbb38d449eb5bccb2702964a52c78811545939712c8280dee0b23b2fa9fbbdd6a0c42476689ace655eba0745b4a21ba108bcd03ad00fdefff416dc74e08486a0538f8fd24989e1b9fc89e675141b7c"}`

	var sig Signature
	err := json.Unmarshal([]byte(signatureJSON), &sig)
	require.NoError(t, err)

	// Check that the method string is lowercased
	require.Equal(t, sig.Method.String(), "rsa")
}

func TestCheckHashes(t *testing.T) {
	var err error
	raw := []byte("Bumblebee")

	// Since only provide an un-supported hash algorithm here,
	// it should be considered as fail.
	unSupported := make(Hashes)
	unSupported["Arthas"] = []byte("is past away.")
	err = CheckHashes(raw, "metaName1", unSupported)
	require.Error(t, err)
	missingMeta, ok := err.(ErrMissingMeta)
	require.True(t, ok)
	require.EqualValues(t, "metaName1", missingMeta.Role)

	// Expected to fail since there is no checksum at all.
	hashes := make(Hashes)
	err = CheckHashes(raw, "metaName2", hashes)
	require.Error(t, err)
	missingMeta, ok = err.(ErrMissingMeta)
	require.True(t, ok)
	require.EqualValues(t, "metaName2", missingMeta.Role)

	// The most standard one.
	hashes[notary.SHA256], err = hex.DecodeString("d13e2b60d74c2e6f4f449b5e536814edf9a4827f5a9f4f957fc92e77609b9c92")
	require.NoError(t, err)
	hashes[notary.SHA512], err = hex.DecodeString("f2330f50d0f3ee56cf0d7f66aad8205e0cb9972c323208ffaa914ef7b3c240ae4774b5bbd1db2ce226ee967cfa9058173a853944f9b44e2e08abca385e2b7ed4")
	require.NoError(t, err)
	err = CheckHashes(raw, "meta", hashes)
	require.NoError(t, err)

	// Expected as success since there are already supported hash here,
	// just ignore the unsupported one.
	hashes["Saar"] = []byte("survives again in CTM.")
	err = CheckHashes(raw, "meta", hashes)
	require.NoError(t, err)

	only256 := make(Hashes)
	only256[notary.SHA256], err = hex.DecodeString("d13e2b60d74c2e6f4f449b5e536814edf9a4827f5a9f4f957fc92e77609b9c92")
	require.NoError(t, err)
	err = CheckHashes(raw, "meta", only256)
	require.NoError(t, err)

	only512 := make(Hashes)
	only512[notary.SHA512], err = hex.DecodeString("f2330f50d0f3ee56cf0d7f66aad8205e0cb9972c323208ffaa914ef7b3c240ae4774b5bbd1db2ce226ee967cfa9058173a853944f9b44e2e08abca385e2b7ed4")
	require.NoError(t, err)
	err = CheckHashes(raw, "meta", only512)
	require.NoError(t, err)

	// Expected to fail due to the failure of sha256
	malicious256 := make(Hashes)
	malicious256[notary.SHA256] = []byte("malicious data")
	err = CheckHashes(raw, "metaName3", malicious256)
	require.Error(t, err)
	badChecksum, ok := err.(ErrMismatchedChecksum)
	require.True(t, ok)
	require.EqualValues(t, ErrMismatchedChecksum{alg: notary.SHA256, name: "metaName3",
		expected: hex.EncodeToString([]byte("malicious data"))}, badChecksum)

	// Expected to fail due to the failure of sha512
	malicious512 := make(Hashes)
	malicious512[notary.SHA512] = []byte("malicious data")
	err = CheckHashes(raw, "metaName4", malicious512)
	require.Error(t, err)
	badChecksum, ok = err.(ErrMismatchedChecksum)
	require.True(t, ok)
	require.EqualValues(t, ErrMismatchedChecksum{alg: notary.SHA512, name: "metaName4",
		expected: hex.EncodeToString([]byte("malicious data"))}, badChecksum)

	// Expected to fail because of the failure of sha512
	// even though the sha256 is OK.
	doubleFace := make(Hashes)
	doubleFace[notary.SHA256], err = hex.DecodeString(
		"d13e2b60d74c2e6f4f449b5e536814edf9a4827f5a9f4f957fc92e77609b9c92")
	require.NoError(t, err)
	doubleFace[notary.SHA512], err = hex.DecodeString(
		"d13e2b60d74c2e6f4f449b5e536814edf9a4827f5a9f4f957fc92e77609b9c92")
	require.NoError(t, err)
	err = CheckHashes(raw, "metaName5", doubleFace)
	require.Error(t, err)
	badChecksum, ok = err.(ErrMismatchedChecksum)
	require.True(t, ok)
	require.EqualValues(t, ErrMismatchedChecksum{alg: notary.SHA512, name: "metaName5",
		expected: "d13e2b60d74c2e6f4f449b5e536814edf9a4827f5a9f4f957fc92e77609b9c92"}, badChecksum)
}

func TestCheckValidHashStructures(t *testing.T) {
	var err error
	hashes := make(Hashes)

	// Expected to fail since there is no checksum at all.
	err = CheckValidHashStructures(hashes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "at least one supported hash needed")

	// Expected to fail even though the checksum of sha384 is valid,
	// because we haven't provided a supported hash algorithm yet (ex: sha256).
	hashes["sha384"], err = hex.DecodeString("64becc3c23843942b1040ffd4743d1368d988ddf046d17d448a6e199c02c3044b425a680112b399d4dbe9b35b7ccc989")
	err = CheckValidHashStructures(hashes)
	require.Error(t, err)
	require.Contains(t, err.Error(), "at least one supported hash needed")

	hashes[notary.SHA256], err = hex.DecodeString("766af0ef090a4f2307e49160fa242db6fb95f071ad81a198eeb7d770e61cd6d8")
	require.NoError(t, err)
	err = CheckValidHashStructures(hashes)
	require.NoError(t, err)

	hashes[notary.SHA512], err = hex.DecodeString("795d9e95db099464b6730844f28effddb010b0d5abae5d5892a6ee04deacb09c9e622f89e816458b5a1a81761278d7d3a6a7c269d9707eff8858b16c51de0315")
	require.NoError(t, err)
	err = CheckValidHashStructures(hashes)
	require.NoError(t, err)

	// Also should be succeed since only check the length of the checksum.
	hashes[notary.SHA256], err = hex.DecodeString("01234567890a4f2307e49160fa242db6fb95f071ad81a198eeb7d770e61cd6d8")
	err = CheckValidHashStructures(hashes)
	require.NoError(t, err)

	// Should failed since the first '0' is missing.
	hashes[notary.SHA256], err = hex.DecodeString("1234567890a4f2307e49160fa242db6fb95f071ad81a198eeb7d770e61cd6d8")
	err = CheckValidHashStructures(hashes)
	require.IsType(t, ErrInvalidChecksum{}, err)
}

func TestCompareMultiHashes(t *testing.T) {
	var err error
	hashes1 := make(Hashes)
	hashes2 := make(Hashes)

	// Expected to fail since there are no checksums at all
	err = CompareMultiHashes(hashes1, hashes2)
	require.Error(t, err)

	// Expected to pass even though the checksum of sha384 isn't a default "supported" hash algorithm valid,
	// because we haven't provided a supported hash algorithm yet (ex: sha256) for the Hashes map to be considered valid
	hashes1["sha384"], err = hex.DecodeString("64becc3c23843942b1040ffd4743d1368d988ddf046d17d448a6e199c02c3044b425a680112b399d4dbe9b35b7ccc989")
	require.NoError(t, err)
	hashes2["sha384"], err = hex.DecodeString("64becc3c23843942b1040ffd4743d1368d988ddf046d17d448a6e199c02c3044b425a680112b399d4dbe9b35b7ccc989")
	require.NoError(t, err)
	err = CompareMultiHashes(hashes1, hashes2)
	require.Error(t, err)

	// Now both have a matching sha256, so this will pass
	hashes1[notary.SHA256], err = hex.DecodeString("766af0ef090a4f2307e49160fa242db6fb95f071ad81a198eeb7d770e61cd6d8")
	require.NoError(t, err)
	hashes2[notary.SHA256], err = hex.DecodeString("766af0ef090a4f2307e49160fa242db6fb95f071ad81a198eeb7d770e61cd6d8")
	require.NoError(t, err)
	err = CompareMultiHashes(hashes1, hashes2)
	require.NoError(t, err)

	// Because the sha384 algorithm isn't a "default hash algorithm", it's still found in the intersection of keys
	// so this check will fail
	hashes2["sha384"], err = hex.DecodeString(strings.Repeat("a", 96))
	require.NoError(t, err)
	err = CompareMultiHashes(hashes1, hashes2)
	require.Error(t, err)
	delete(hashes2, "sha384")

	// only add a sha512 to hashes1, but comparison will still succeed because there's no mismatch and we have the sha256 match
	hashes1[notary.SHA512], err = hex.DecodeString("795d9e95db099464b6730844f28effddb010b0d5abae5d5892a6ee04deacb09c9e622f89e816458b5a1a81761278d7d3a6a7c269d9707eff8858b16c51de0315")
	require.NoError(t, err)
	err = CompareMultiHashes(hashes1, hashes2)
	require.NoError(t, err)

	// remove sha256 from hashes1, comparison will fail now because there are no matches
	delete(hashes1, notary.SHA256)
	err = CompareMultiHashes(hashes1, hashes2)
	require.Error(t, err)

	// add sha512 to hashes2, comparison will now pass because both have matching sha512s
	hashes2[notary.SHA512], err = hex.DecodeString("795d9e95db099464b6730844f28effddb010b0d5abae5d5892a6ee04deacb09c9e622f89e816458b5a1a81761278d7d3a6a7c269d9707eff8858b16c51de0315")
	require.NoError(t, err)
	err = CompareMultiHashes(hashes1, hashes2)
	require.NoError(t, err)

	// change the sha512 for hashes2, comparison will now fail
	hashes2[notary.SHA512], err = hex.DecodeString(strings.Repeat("a", notary.SHA512HexSize))
	require.NoError(t, err)
	err = CompareMultiHashes(hashes1, hashes2)
	require.Error(t, err)
}

func TestFileMetaEquals(t *testing.T) {
	var err error

	hashes1 := make(Hashes)
	hashes2 := make(Hashes)
	hashes1["sha384"], err = hex.DecodeString("64becc3c23843942b1040ffd4743d1368d988ddf046d17d448a6e199c02c3044b425a680112b399d4dbe9b35b7ccc989")
	require.NoError(t, err)
	hashes2[notary.SHA256], err = hex.DecodeString("766af0ef090a4f2307e49160fa242db6fb95f071ad81a198eeb7d770e61cd6d8")
	require.NoError(t, err)

	rawMessage1, rawMessage2 := json.RawMessage{}, json.RawMessage{}
	require.NoError(t, rawMessage1.UnmarshalJSON([]byte("hello")))
	require.NoError(t, rawMessage2.UnmarshalJSON([]byte("there")))

	f1 := []FileMeta{
		{
			Length: 1,
			Hashes: hashes1,
		},
		{
			Length: 2,
			Hashes: hashes1,
		},
		{
			Length: 1,
			Hashes: hashes2,
		},
		{
			Length: 1,
			Hashes: hashes1,
			Custom: &rawMessage1,
		},
		{},
	}

	f2 := []FileMeta{
		{
			Length: 1,
			Hashes: hashes1,
			Custom: &rawMessage1,
		},
		{
			Length: 1,
			Hashes: hashes1,
			Custom: &rawMessage2,
		},
		{
			Length: 1,
			Hashes: hashes1,
		},
	}

	require.False(t, FileMeta{}.Equals(f1[0]))
	require.True(t, f1[0].Equals(f1[0]))
	for _, meta := range f1[1:] {
		require.False(t, f1[0].Equals(meta))
	}
	require.True(t, f2[0].Equals(f2[0]))
	for _, meta := range f2[1:] {
		require.False(t, f2[0].Equals(meta))
	}
}
