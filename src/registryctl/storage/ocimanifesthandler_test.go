package storage

import (
	"context"
	"regexp"
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/manifest/ocischema"
	"github.com/goharbor/harbor/src/registryctl/storage/driver/inmemory"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

func TestVerifyOCIManifestNonDistributableLayer(t *testing.T) {
	ctx := context.Background()
	inmemoryDriver := inmemory.New()
	registry := createRegistry(t, inmemoryDriver,
		ManifestURLsAllowRegexp(regexp.MustCompile("^https?://foo")),
		ManifestURLsDenyRegexp(regexp.MustCompile("^https?://foo/nope")))
	repo := makeRepository(t, registry, "test")
	manifestService := makeManifestService(t, repo)

	config, err := repo.Blobs(ctx).Put(ctx, v1.MediaTypeImageConfig, nil)
	if err != nil {
		t.Fatal(err)
	}

	layer, err := repo.Blobs(ctx).Put(ctx, v1.MediaTypeImageLayerGzip, nil)
	if err != nil {
		t.Fatal(err)
	}

	nonDistributableLayer := distribution.Descriptor{
		Digest:    "sha256:463435349086340864309863409683460843608348608934092322395278926a",
		Size:      6323,
		MediaType: v1.MediaTypeImageLayerNonDistributableGzip,
	}

	template := ocischema.Manifest{
		Versioned: manifest.Versioned{
			SchemaVersion: 2,
			MediaType:     v1.MediaTypeImageManifest,
		},
		Config: config,
	}

	type testcase struct {
		BaseLayer distribution.Descriptor
		URLs      []string
		Err       error
	}

	cases := []testcase{
		{
			nonDistributableLayer,
			nil,
			distribution.ErrManifestBlobUnknown{Digest: nonDistributableLayer.Digest},
		},
		{
			layer,
			[]string{"http://foo/bar"},
			nil,
		},
		{
			nonDistributableLayer,
			[]string{"file:///local/file"},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{"http://foo/bar#baz"},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{""},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{"https://foo/bar", ""},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{"", "https://foo/bar"},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{"http://nope/bar"},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{"http://foo/nope"},
			errInvalidURL,
		},
		{
			nonDistributableLayer,
			[]string{"http://foo/bar"},
			nil,
		},
		{
			nonDistributableLayer,
			[]string{"https://foo/bar"},
			nil,
		},
	}

	for _, c := range cases {
		m := template
		l := c.BaseLayer
		l.URLs = c.URLs
		m.Layers = []distribution.Descriptor{l}
		dm, err := ocischema.FromStruct(m)
		if err != nil {
			t.Error(err)
			continue
		}

		_, err = manifestService.Put(ctx, dm)
		if verr, ok := err.(distribution.ErrManifestVerification); ok {
			// Extract the first error
			if len(verr) == 2 {
				if _, ok = verr[1].(distribution.ErrManifestBlobUnknown); ok {
					err = verr[0]
				}
			} else if len(verr) == 1 {
				err = verr[0]
			}
		}
		if err != c.Err {
			t.Errorf("%#v: expected %v, got %v", l, c.Err, err)
		}
	}
}
