package utils

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/log"

	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
)

func Retag(srcImage, destImage *models.Image) error {
	srcClient, err := NewRepositoryClientForUI("harbor-ui", getRepoName(srcImage))
	if err != nil {
		return err
	}
	destClient := srcClient
	if getRepoName(srcImage) != getRepoName(destImage) {
		destClient, err = NewRepositoryClientForUI("harbor-ui", getRepoName(destImage))
		if err != nil {
			return err
		}
	}

	accepted := []string{schema1.MediaTypeManifest, schema2.MediaTypeManifest}
	digest, mediaType, payload, err := srcClient.PullManifest(srcImage.Tag, accepted)
	if err != nil {
		return err
	}

	manifest, _, err := registry.UnMarshal(mediaType, payload)
	if err != nil {
		return err
	}

	destDigest, exist, err := destClient.ManifestExist(destImage.Tag)
	if err != nil {
		log.Errorf("check existence of manifest '%s:%s' error: %v", destClient.Name, destImage.Tag, err)
		return err
	}
	if exist && destDigest == digest {
		log.Infof("manifest of '%s:%s' already exist", destClient.Name, destImage.Tag)
		return nil
	}

	if getRepoName(srcImage) != getRepoName(destImage) {
		for _, descriptor := range manifest.References() {
			err := destClient.MountBlob(descriptor.Digest.String(), srcClient.Name)
			if err != nil {
				log.Errorf("mount blob '%s' error: %v", descriptor.Digest.String(), err)
				return err
			}
		}
	}

	if _, err = destClient.PushManifest(destImage.Tag, mediaType, payload); err != nil {
		log.Errorf("push manifest '%s:%s' error: %v", destClient.Name, destImage.Tag, err)
		return err
	}

	return nil
}

func getRepoName(image *models.Image) string {
	return fmt.Sprintf("%s/%s", image.Project, image.Repo)
}