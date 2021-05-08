package tencentcr

import (
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/filter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
)

var (
	tcrQPSLimit = 10
)

/**
	* Implement ArtifactRegistry Interface
**/
var _ adp.ArtifactRegistry = &adapter{}

func filterToPatterns(filters []*model.Filter) (namespacePattern, repoPattern, tagsPattern string) {
	for _, filter := range filters {
		if filter.Type == model.FilterTypeName {
			repoPattern = filter.Value.(string)
		}
		if filter.Type == model.FilterTypeTag {
			tagsPattern = filter.Value.(string)
		}
	}
	namespacePattern = strings.Split(repoPattern, "/")[0]
	return
}

func (a *adapter) FetchArtifacts(filters []*model.Filter) (resources []*model.Resource, err error) {
	// get filter pattern
	var namespacePattern, repoPattern, tagsPattern = filterToPatterns(filters)
	log.Debugf("[tencent-tcr.FetchArtifacts] namespacePattern=%s repoPattern=%s tagsPattern=%s", namespacePattern, repoPattern, tagsPattern)

	// 1. list namespaces
	var namespaces []string
	namespaces, err = a.listCandidateNamespaces(namespacePattern)
	if err != nil {
		return
	}
	log.Debugf("[tencent-tcr.FetchArtifacts] namespaces=%v", namespaces)

	// 2. list repos
	// var filteredRepos []tcr.TcrRepositoryInfo
	var repos []*model.Repository
	var repositories []*model.Repository
	for _, ns := range namespaces {
		tcrRepos, err := a.listReposByNamespace(ns)
		if err != nil {
			return nil, err
		}

		if len(tcrRepos) == 0 {
			continue
		}
		for _, tcrRepo := range tcrRepos {
			repositories = append(repositories, &model.Repository{
				Name: *tcrRepo.Name,
			})
		}
	}
	repos, _ = filter.DoFilterRepositories(repositories, filters)
	log.Debugf("[tencent-tcr.FetchArtifacts] filteredRepos=%d", len(repos))

	// 4. list images
	var rawResources = make([]*model.Resource, len(repos))
	runner := utils.NewLimitedConcurrentRunner(tcrQPSLimit)

	for i, r := range repos {
		// !copy
		index := i
		repo := r

		runner.AddTask(func() error {
			var images []string
			repoArr := strings.Split(repo.Name, "/")
			_, images, err = a.getImages(repoArr[0], strings.Join(repoArr[1:], "/"), "")
			if err != nil {
				return fmt.Errorf("[tencent-tcr.FetchArtifacts.listImages] runner=%d repo=%s, error=%v", index, repo.Name, err)
			}

			var filteredImages []string
			if tagsPattern != "" {
				for _, image := range images {
					var ok bool
					ok, err = util.Match(tagsPattern, image)
					if err != nil {
						return fmt.Errorf("[tencent-tcr.FetchArtifacts.matchImage] image='%s', error=%v", image, err)
					}
					if ok {
						filteredImages = append(filteredImages, image)
					}
				}
			} else {
				filteredImages = images
			}

			log.Debugf("[tencent-tcr.FetchArtifacts] repo=%s, images=%v, filteredImages=%v", repo.Name, images, filteredImages)

			if len(filteredImages) > 0 {
				rawResources[index] = &model.Resource{
					Type:     model.ResourceTypeImage,
					Registry: a.registry,
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{
							Name: repo.Name,
						},
						Vtags: filteredImages,
					},
				}
			}

			return nil
		})
	}
	if err = runner.Wait(); err != nil {
		return nil, fmt.Errorf("failed to fetch artifacts: %v", err)
	}

	for _, res := range rawResources {
		if res != nil {
			resources = append(resources, res)
		}
	}
	log.Debugf("[tencent-tcr.FetchArtifacts] resources.size=%d", len(resources))

	return
}

func (a *adapter) listCandidateNamespaces(namespacePattern string) (namespaces []string, err error) {
	// filter namespaces
	if len(namespacePattern) > 0 {
		if nms, ok := util.IsSpecificPathComponent(namespacePattern); ok {
			// Check is exist
			var exist bool
			for _, ns := range nms {
				exist, err = a.isNamespaceExist(ns)
				if err != nil {
					return
				}
				if !exist {
					continue
				}
				namespaces = append(namespaces, nms...)
			}
		}
	}

	if len(namespaces) > 0 {
		log.Debugf("[tencent-tcr.listCandidateNamespaces] pattern=%s, namespaces=%v", namespacePattern, namespaces)
		return namespaces, nil
	}

	// list all
	return a.listNamespaces()
}

func (a *adapter) DeleteManifest(repository, reference string) (err error) {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return fmt.Errorf("tcr only support repo in format <namespace>/<name>, but got: %s", repository)
	}
	log.Warningf("[tencent-tcr.DeleteManifest] namespace=%s, repository=%s, tag=%s", parts[0], parts[1], reference)

	return a.deleteImage(parts[0], parts[1], reference)
}
