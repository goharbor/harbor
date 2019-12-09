package storage

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/ocischema"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/storage/cache/memory"
	"github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/inmemory"
	"github.com/docker/distribution/testutil"
	"github.com/docker/libtrust"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

type manifestStoreTestEnv struct {
	ctx        context.Context
	driver     driver.StorageDriver
	registry   distribution.Namespace
	repository distribution.Repository
	name       reference.Named
	tag        string
}

func newManifestStoreTestEnv(t *testing.T, name reference.Named, tag string, options ...RegistryOption) *manifestStoreTestEnv {
	ctx := context.Background()
	driver := inmemory.New()
	registry, err := NewRegistry(ctx, driver, options...)
	if err != nil {
		t.Fatalf("error creating registry: %v", err)
	}

	repo, err := registry.Repository(ctx, name)
	if err != nil {
		t.Fatalf("unexpected error getting repo: %v", err)
	}

	return &manifestStoreTestEnv{
		ctx:        ctx,
		driver:     driver,
		registry:   registry,
		repository: repo,
		name:       name,
		tag:        tag,
	}
}

func TestManifestStorage(t *testing.T) {
	k, err := libtrust.GenerateECP256PrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	testManifestStorage(t, true, BlobDescriptorCacheProvider(memory.NewInMemoryBlobDescriptorCacheProvider()), EnableDelete, EnableRedirect, Schema1SigningKey(k), EnableSchema1)
}

func TestManifestStorageV1Unsupported(t *testing.T) {
	k, err := libtrust.GenerateECP256PrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	testManifestStorage(t, false, BlobDescriptorCacheProvider(memory.NewInMemoryBlobDescriptorCacheProvider()), EnableDelete, EnableRedirect, Schema1SigningKey(k))
}

func testManifestStorage(t *testing.T, schema1Enabled bool, options ...RegistryOption) {
	repoName, _ := reference.WithName("foo/bar")
	env := newManifestStoreTestEnv(t, repoName, "thetag", options...)
	ctx := context.Background()
	ms, err := env.repository.Manifests(ctx)
	if err != nil {
		t.Fatal(err)
	}

	m := schema1.Manifest{
		Versioned: manifest.Versioned{
			SchemaVersion: 1,
		},
		Name: env.name.Name(),
		Tag:  env.tag,
	}

	// Build up some test layers and add them to the manifest, saving the
	// readseekers for upload later.
	testLayers := map[digest.Digest]io.ReadSeeker{}
	for i := 0; i < 2; i++ {
		rs, ds, err := testutil.CreateRandomTarFile()
		if err != nil {
			t.Fatalf("unexpected error generating test layer file")
		}
		dgst := digest.Digest(ds)

		testLayers[digest.Digest(dgst)] = rs
		m.FSLayers = append(m.FSLayers, schema1.FSLayer{
			BlobSum: dgst,
		})
		m.History = append(m.History, schema1.History{
			V1Compatibility: "",
		})

	}

	pk, err := libtrust.GenerateECP256PrivateKey()
	if err != nil {
		t.Fatalf("unexpected error generating private key: %v", err)
	}

	sm, merr := schema1.Sign(&m, pk)
	if merr != nil {
		t.Fatalf("error signing manifest: %v", err)
	}

	_, err = ms.Put(ctx, sm)
	if err == nil {
		t.Fatalf("expected errors putting manifest with full verification")
	}

	// If schema1 is not enabled, do a short version of this test, just checking
	// if we get the right error when we Put
	if !schema1Enabled {
		if err != distribution.ErrSchemaV1Unsupported {
			t.Fatalf("got the wrong error when schema1 is disabled: %s", err)
		}
		return
	}

	switch err := err.(type) {
	case distribution.ErrManifestVerification:
		if len(err) != 2 {
			t.Fatalf("expected 2 verification errors: %#v", err)
		}

		for _, err := range err {
			if _, ok := err.(distribution.ErrManifestBlobUnknown); !ok {
				t.Fatalf("unexpected error type: %v", err)
			}
		}
	default:
		t.Fatalf("unexpected error verifying manifest: %v", err)
	}

	// Now, upload the layers that were missing!
	for dgst, rs := range testLayers {
		wr, err := env.repository.Blobs(env.ctx).Create(env.ctx)
		if err != nil {
			t.Fatalf("unexpected error creating test upload: %v", err)
		}

		if _, err := io.Copy(wr, rs); err != nil {
			t.Fatalf("unexpected error copying to upload: %v", err)
		}

		if _, err := wr.Commit(env.ctx, distribution.Descriptor{Digest: dgst}); err != nil {
			t.Fatalf("unexpected error finishing upload: %v", err)
		}
	}

	var manifestDigest digest.Digest
	if manifestDigest, err = ms.Put(ctx, sm); err != nil {
		t.Fatalf("unexpected error putting manifest: %v", err)
	}

	exists, err := ms.Exists(ctx, manifestDigest)
	if err != nil {
		t.Fatalf("unexpected error checking manifest existence: %#v", err)
	}

	if !exists {
		t.Fatalf("manifest should exist")
	}

	fromStore, err := ms.Get(ctx, manifestDigest)
	if err != nil {
		t.Fatalf("unexpected error fetching manifest: %v", err)
	}

	fetchedManifest, ok := fromStore.(*schema1.SignedManifest)
	if !ok {
		t.Fatalf("unexpected manifest type from signedstore")
	}

	if !bytes.Equal(fetchedManifest.Canonical, sm.Canonical) {
		t.Fatalf("fetched payload does not match original payload: %q != %q", fetchedManifest.Canonical, sm.Canonical)
	}

	_, pl, err := fetchedManifest.Payload()
	if err != nil {
		t.Fatalf("error getting payload %#v", err)
	}

	fetchedJWS, err := libtrust.ParsePrettySignature(pl, "signatures")
	if err != nil {
		t.Fatalf("unexpected error parsing jws: %v", err)
	}

	payload, err := fetchedJWS.Payload()
	if err != nil {
		t.Fatalf("unexpected error extracting payload: %v", err)
	}

	// Now that we have a payload, take a moment to check that the manifest is
	// return by the payload digest.

	dgst := digest.FromBytes(payload)
	exists, err = ms.Exists(ctx, dgst)
	if err != nil {
		t.Fatalf("error checking manifest existence by digest: %v", err)
	}

	if !exists {
		t.Fatalf("manifest %s should exist", dgst)
	}

	fetchedByDigest, err := ms.Get(ctx, dgst)
	if err != nil {
		t.Fatalf("unexpected error fetching manifest by digest: %v", err)
	}

	byDigestManifest, ok := fetchedByDigest.(*schema1.SignedManifest)
	if !ok {
		t.Fatalf("unexpected manifest type from signedstore")
	}

	if !bytes.Equal(byDigestManifest.Canonical, fetchedManifest.Canonical) {
		t.Fatalf("fetched manifest not equal: %q != %q", byDigestManifest.Canonical, fetchedManifest.Canonical)
	}

	sigs, err := fetchedJWS.Signatures()
	if err != nil {
		t.Fatalf("unable to extract signatures: %v", err)
	}

	if len(sigs) != 1 {
		t.Fatalf("unexpected number of signatures: %d != %d", len(sigs), 1)
	}

	// Now, push the same manifest with a different key
	pk2, err := libtrust.GenerateECP256PrivateKey()
	if err != nil {
		t.Fatalf("unexpected error generating private key: %v", err)
	}

	sm2, err := schema1.Sign(&m, pk2)
	if err != nil {
		t.Fatalf("unexpected error signing manifest: %v", err)
	}
	_, pl, err = sm2.Payload()
	if err != nil {
		t.Fatalf("error getting payload %#v", err)
	}

	jws2, err := libtrust.ParsePrettySignature(pl, "signatures")
	if err != nil {
		t.Fatalf("error parsing signature: %v", err)
	}

	sigs2, err := jws2.Signatures()
	if err != nil {
		t.Fatalf("unable to extract signatures: %v", err)
	}

	if len(sigs2) != 1 {
		t.Fatalf("unexpected number of signatures: %d != %d", len(sigs2), 1)
	}

	if manifestDigest, err = ms.Put(ctx, sm2); err != nil {
		t.Fatalf("unexpected error putting manifest: %v", err)
	}

	fromStore, err = ms.Get(ctx, manifestDigest)
	if err != nil {
		t.Fatalf("unexpected error fetching manifest: %v", err)
	}

	fetched, ok := fromStore.(*schema1.SignedManifest)
	if !ok {
		t.Fatalf("unexpected type from signed manifeststore : %T", fetched)
	}

	if _, err := schema1.Verify(fetched); err != nil {
		t.Fatalf("unexpected error verifying manifest: %v", err)
	}

	_, pl, err = fetched.Payload()
	if err != nil {
		t.Fatalf("error getting payload %#v", err)
	}

	receivedJWS, err := libtrust.ParsePrettySignature(pl, "signatures")
	if err != nil {
		t.Fatalf("unexpected error parsing jws: %v", err)
	}

	receivedPayload, err := receivedJWS.Payload()
	if err != nil {
		t.Fatalf("unexpected error extracting received payload: %v", err)
	}

	if !bytes.Equal(receivedPayload, payload) {
		t.Fatalf("payloads are not equal")
	}

	// Test deleting manifests
	err = ms.Delete(ctx, dgst)
	if err != nil {
		t.Fatalf("unexpected an error deleting manifest by digest: %v", err)
	}

	exists, err = ms.Exists(ctx, dgst)
	if err != nil {
		t.Fatalf("Error querying manifest existence")
	}
	if exists {
		t.Errorf("Deleted manifest should not exist")
	}

	deletedManifest, err := ms.Get(ctx, dgst)
	if err == nil {
		t.Errorf("Unexpected success getting deleted manifest")
	}
	switch err.(type) {
	case distribution.ErrManifestUnknownRevision:
		break
	default:
		t.Errorf("Unexpected error getting deleted manifest: %s", reflect.ValueOf(err).Type())
	}

	if deletedManifest != nil {
		t.Errorf("Deleted manifest get returned non-nil")
	}

	// Re-upload should restore manifest to a good state
	_, err = ms.Put(ctx, sm)
	if err != nil {
		t.Errorf("Error re-uploading deleted manifest")
	}

	exists, err = ms.Exists(ctx, dgst)
	if err != nil {
		t.Fatalf("Error querying manifest existence")
	}
	if !exists {
		t.Errorf("Restored manifest should exist")
	}

	deletedManifest, err = ms.Get(ctx, dgst)
	if err != nil {
		t.Errorf("Unexpected error getting manifest")
	}
	if deletedManifest == nil {
		t.Errorf("Deleted manifest get returned non-nil")
	}

	r, err := NewRegistry(ctx, env.driver, BlobDescriptorCacheProvider(memory.NewInMemoryBlobDescriptorCacheProvider()), EnableRedirect)
	if err != nil {
		t.Fatalf("error creating registry: %v", err)
	}
	repo, err := r.Repository(ctx, env.name)
	if err != nil {
		t.Fatalf("unexpected error getting repo: %v", err)
	}
	ms, err = repo.Manifests(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = ms.Delete(ctx, dgst)
	if err == nil {
		t.Errorf("Unexpected success deleting while disabled")
	}
}

func TestOCIManifestStorage(t *testing.T) {
	testOCIManifestStorage(t, "includeMediaTypes=true", true)
	testOCIManifestStorage(t, "includeMediaTypes=false", false)
}

func testOCIManifestStorage(t *testing.T, testname string, includeMediaTypes bool) {
	var imageMediaType string
	var indexMediaType string
	if includeMediaTypes {
		imageMediaType = v1.MediaTypeImageManifest
		indexMediaType = v1.MediaTypeImageIndex
	} else {
		imageMediaType = ""
		indexMediaType = ""
	}

	repoName, _ := reference.WithName("foo/bar")
	env := newManifestStoreTestEnv(t, repoName, "thetag",
		BlobDescriptorCacheProvider(memory.NewInMemoryBlobDescriptorCacheProvider()),
		EnableDelete, EnableRedirect)

	ctx := context.Background()
	ms, err := env.repository.Manifests(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Build a manifest and store it and its layers in the registry

	blobStore := env.repository.Blobs(ctx)
	builder := ocischema.NewManifestBuilder(blobStore, []byte{}, map[string]string{})
	err = builder.(*ocischema.Builder).SetMediaType(imageMediaType)
	if err != nil {
		t.Fatal(err)
	}

	// Add some layers
	for i := 0; i < 2; i++ {
		rs, ds, err := testutil.CreateRandomTarFile()
		if err != nil {
			t.Fatalf("%s: unexpected error generating test layer file", testname)
		}
		dgst := digest.Digest(ds)

		wr, err := env.repository.Blobs(env.ctx).Create(env.ctx)
		if err != nil {
			t.Fatalf("%s: unexpected error creating test upload: %v", testname, err)
		}

		if _, err := io.Copy(wr, rs); err != nil {
			t.Fatalf("%s: unexpected error copying to upload: %v", testname, err)
		}

		if _, err := wr.Commit(env.ctx, distribution.Descriptor{Digest: dgst}); err != nil {
			t.Fatalf("%s: unexpected error finishing upload: %v", testname, err)
		}

		builder.AppendReference(distribution.Descriptor{Digest: dgst})
	}

	manifest, err := builder.Build(ctx)
	if err != nil {
		t.Fatalf("%s: unexpected error generating manifest: %v", testname, err)
	}

	// before putting the manifest test for proper handling of SchemaVersion

	if manifest.(*ocischema.DeserializedManifest).Manifest.SchemaVersion != 2 {
		t.Fatalf("%s: unexpected error generating default version for oci manifest", testname)
	}
	manifest.(*ocischema.DeserializedManifest).Manifest.SchemaVersion = 0

	var manifestDigest digest.Digest
	if manifestDigest, err = ms.Put(ctx, manifest); err != nil {
		if err.Error() != "unrecognized manifest schema version 0" {
			t.Fatalf("%s: unexpected error putting manifest: %v", testname, err)
		}
		manifest.(*ocischema.DeserializedManifest).Manifest.SchemaVersion = 2
		if manifestDigest, err = ms.Put(ctx, manifest); err != nil {
			t.Fatalf("%s: unexpected error putting manifest: %v", testname, err)
		}
	}

	// Also create an image index that contains the manifest

	descriptor, err := env.registry.BlobStatter().Stat(ctx, manifestDigest)
	if err != nil {
		t.Fatalf("%s: unexpected error getting manifest descriptor", testname)
	}
	descriptor.MediaType = v1.MediaTypeImageManifest

	platformSpec := manifestlist.PlatformSpec{
		Architecture: "atari2600",
		OS:           "CP/M",
	}

	manifestDescriptors := []manifestlist.ManifestDescriptor{
		{
			Descriptor: descriptor,
			Platform:   platformSpec,
		},
	}

	imageIndex, err := manifestlist.FromDescriptorsWithMediaType(manifestDescriptors, indexMediaType)
	if err != nil {
		t.Fatalf("%s: unexpected error creating image index: %v", testname, err)
	}

	var indexDigest digest.Digest
	if indexDigest, err = ms.Put(ctx, imageIndex); err != nil {
		t.Fatalf("%s: unexpected error putting image index: %v", testname, err)
	}

	// Now check that we can retrieve the manifest

	fromStore, err := ms.Get(ctx, manifestDigest)
	if err != nil {
		t.Fatalf("%s: unexpected error fetching manifest: %v", testname, err)
	}

	fetchedManifest, ok := fromStore.(*ocischema.DeserializedManifest)
	if !ok {
		t.Fatalf("%s: unexpected type for fetched manifest", testname)
	}

	if fetchedManifest.MediaType != imageMediaType {
		t.Fatalf("%s: unexpected MediaType for result, %s", testname, fetchedManifest.MediaType)
	}

	if fetchedManifest.SchemaVersion != ocischema.SchemaVersion.SchemaVersion {
		t.Fatalf("%s: unexpected schema version for result, %d", testname, fetchedManifest.SchemaVersion)
	}

	payloadMediaType, _, err := fromStore.Payload()
	if err != nil {
		t.Fatalf("%s: error getting payload %v", testname, err)
	}

	if payloadMediaType != v1.MediaTypeImageManifest {
		t.Fatalf("%s: unexpected MediaType for manifest payload, %s", testname, payloadMediaType)
	}

	// and the image index

	fromStore, err = ms.Get(ctx, indexDigest)
	if err != nil {
		t.Fatalf("%s: unexpected error fetching image index: %v", testname, err)
	}

	fetchedIndex, ok := fromStore.(*manifestlist.DeserializedManifestList)
	if !ok {
		t.Fatalf("%s: unexpected type for fetched manifest", testname)
	}

	if fetchedIndex.MediaType != indexMediaType {
		t.Fatalf("%s: unexpected MediaType for result, %s", testname, fetchedManifest.MediaType)
	}

	payloadMediaType, _, err = fromStore.Payload()
	if err != nil {
		t.Fatalf("%s: error getting payload %v", testname, err)
	}

	if payloadMediaType != v1.MediaTypeImageIndex {
		t.Fatalf("%s: unexpected MediaType for index payload, %s", testname, payloadMediaType)
	}

}

// TestLinkPathFuncs ensures that the link path functions behavior are locked
// down and implemented as expected.
func TestLinkPathFuncs(t *testing.T) {
	for _, testcase := range []struct {
		repo       string
		digest     digest.Digest
		linkPathFn linkPathFunc
		expected   string
	}{
		{
			repo:       "foo/bar",
			digest:     "sha256:deadbeaf98fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			linkPathFn: blobLinkPath,
			expected:   "/docker/registry/v2/repositories/foo/bar/_layers/sha256/deadbeaf98fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855/link",
		},
		{
			repo:       "foo/bar",
			digest:     "sha256:deadbeaf98fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			linkPathFn: manifestRevisionLinkPath,
			expected:   "/docker/registry/v2/repositories/foo/bar/_manifests/revisions/sha256/deadbeaf98fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855/link",
		},
	} {
		p, err := testcase.linkPathFn(testcase.repo, testcase.digest)
		if err != nil {
			t.Fatalf("unexpected error calling linkPathFn(pm, %q, %q): %v", testcase.repo, testcase.digest, err)
		}

		if p != testcase.expected {
			t.Fatalf("incorrect path returned: %q != %q", p, testcase.expected)
		}
	}
}
