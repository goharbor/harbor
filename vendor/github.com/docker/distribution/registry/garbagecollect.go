package registry

import (
	"fmt"
	"os"

	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/storage"
	"github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/distribution/registry/storage/driver/factory"

	"github.com/spf13/cobra"
)

func markAndSweep(storageDriver driver.StorageDriver) error {
	ctx := context.Background()

	// Construct a registry
	registry, err := storage.NewRegistry(ctx, storageDriver)
	if err != nil {
		return fmt.Errorf("failed to construct registry: %v", err)
	}

	repositoryEnumerator, ok := registry.(distribution.RepositoryEnumerator)
	if !ok {
		return fmt.Errorf("coercion error: unable to convert Namespace to RepositoryEnumerator")
	}

	// mark
	markSet := make(map[digest.Digest]struct{})
	err = repositoryEnumerator.Enumerate(ctx, func(repoName string) error {
		var err error
		named, err := reference.ParseNamed(repoName)
		if err != nil {
			return fmt.Errorf("failed to parse repo name %s: %v", repoName, err)
		}
		repository, err := registry.Repository(ctx, named)
		if err != nil {
			return fmt.Errorf("failed to construct repository: %v", err)
		}

		manifestService, err := repository.Manifests(ctx)
		if err != nil {
			return fmt.Errorf("failed to construct manifest service: %v", err)
		}

		manifestEnumerator, ok := manifestService.(distribution.ManifestEnumerator)
		if !ok {
			return fmt.Errorf("coercion error: unable to convert ManifestService into ManifestEnumerator")
		}

		err = manifestEnumerator.Enumerate(ctx, func(dgst digest.Digest) error {
			// Mark the manifest's blob
			markSet[dgst] = struct{}{}

			manifest, err := manifestService.Get(ctx, dgst)
			if err != nil {
				return fmt.Errorf("failed to retrieve manifest for digest %v: %v", dgst, err)
			}

			descriptors := manifest.References()
			for _, descriptor := range descriptors {
				markSet[descriptor.Digest] = struct{}{}
			}

			switch manifest.(type) {
			case *schema1.SignedManifest:
				signaturesGetter, ok := manifestService.(distribution.SignaturesGetter)
				if !ok {
					return fmt.Errorf("coercion error: unable to convert ManifestSErvice into SignaturesGetter")
				}
				signatures, err := signaturesGetter.GetSignatures(ctx, dgst)
				if err != nil {
					return fmt.Errorf("failed to get signatures for signed manifest: %v", err)
				}
				for _, signatureDigest := range signatures {
					markSet[signatureDigest] = struct{}{}
				}
				break
			case *schema2.DeserializedManifest:
				config := manifest.(*schema2.DeserializedManifest).Config
				markSet[config.Digest] = struct{}{}
				break
			}

			return nil
		})

		return err
	})

	if err != nil {
		return fmt.Errorf("failed to mark: %v\n", err)
	}

	// sweep
	blobService := registry.Blobs()
	deleteSet := make(map[digest.Digest]struct{})
	err = blobService.Enumerate(ctx, func(dgst digest.Digest) error {
		// check if digest is in markSet. If not, delete it!
		if _, ok := markSet[dgst]; !ok {
			deleteSet[dgst] = struct{}{}
		}
		return nil
	})

	// Construct vacuum
	vacuum := storage.NewVacuum(ctx, storageDriver)
	for dgst := range deleteSet {
		err = vacuum.RemoveBlob(string(dgst))
		if err != nil {
			return fmt.Errorf("failed to delete blob %s: %v\n", dgst, err)
		}
	}

	return err
}

// GCCmd is the cobra command that corresponds to the garbage-collect subcommand
var GCCmd = &cobra.Command{
	Use:   "garbage-collect <config>",
	Short: "`garbage-collects` deletes layers not referenced by any manifests",
	Long:  "`garbage-collects` deletes layers not referenced by any manifests",
	Run: func(cmd *cobra.Command, args []string) {

		config, err := resolveConfiguration(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
			cmd.Usage()
			os.Exit(1)
		}

		driver, err := factory.Create(config.Storage.Type(), config.Storage.Parameters())
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to construct %s driver: %v", config.Storage.Type(), err)
			os.Exit(1)
		}

		err = markAndSweep(driver)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to garbage collect: %v", err)
			os.Exit(1)
		}
	},
}
