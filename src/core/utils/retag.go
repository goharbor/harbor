// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry"

	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
)

// Retag tags an image to another
func Retag(srcImage, destImage *models.Image) error {
	isSameRepo := getRepoName(srcImage) == getRepoName(destImage)
	srcClient, err := NewRepositoryClientForUIWithMiddleware("harbor-ui", getRepoName(srcImage))
	if err != nil {
		return err
	}
	destClient := srcClient
	if !isSameRepo {
		destClient, err = NewRepositoryClientForUIWithMiddleware("harbor-ui", getRepoName(destImage))
		if err != nil {
			return err
		}
	}

	_, exist, err := srcClient.ManifestExist(srcImage.Tag)
	if err != nil {
		log.Errorf("check existence of manifest '%s:%s' error: %v", srcClient.Name, srcImage.Tag, err)
		return err
	}

	if !exist {
		log.Errorf("source image %s:%s not found", srcClient.Name, srcImage.Tag)
		return fmt.Errorf("image %s:%s not found", srcClient.Name, srcImage.Tag)
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

	if !isSameRepo {
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
