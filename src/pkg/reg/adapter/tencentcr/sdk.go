package tencentcr

import (
	"errors"
	"strings"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tcr/v20190924"
)

func (a *adapter) createPrivateNamespace(namespace string) (err error) {
	if a.tcrClient == nil {
		err = errors.New("[tencent-tcr.createPrivateNamespace] nil tcr client")
		return
	}

	// 1. if exist skip
	log.Debugf("[tencent-tcr.PrepareForPush.createPrivateNamespace] namespace=%s", namespace)
	var exist bool
	exist, err = a.isNamespaceExist(namespace)
	if err != nil {
		return
	}
	if exist {
		log.Warningf("[tencent-tcr.PrepareForPush.createPrivateNamespace.skip_exist] namespace=%s", namespace)
		return
	}

	// !!! 2. WARNING: for safety, auto create namespace is private.
	var req = tcr.NewCreateNamespaceRequest()
	req.NamespaceName = &namespace
	req.RegistryId = a.registryID
	var isPublic = false
	req.IsPublic = &isPublic
	tcr.NewCreateNamespaceResponse()
	_, err = a.tcrClient.CreateNamespace(req)
	if err != nil {
		log.Debugf("[tencent-tcr.PrepareForPush.createPrivateNamespace] error=%v", err)
		return
	}
	return
}

func (a *adapter) createRepository(namespace, repository string) (err error) {
	if a.tcrClient == nil {
		err = errors.New("[tencent-tcr.createRepository] nil tcr client")
		return
	}

	// 1. if exist skip
	log.Debugf("[tencent-tcr.PrepareForPush.createRepository] namespace=%s, repository=%s", namespace, repository)
	var repoReq = tcr.NewDescribeRepositoriesRequest()
	repoReq.RegistryId = a.registryID
	repoReq.NamespaceName = &namespace
	repoReq.RepositoryName = &repository
	var repoResp = tcr.NewDescribeRepositoriesResponse()
	repoResp, err = a.tcrClient.DescribeRepositories(repoReq)
	if err != nil {
		return
	}
	if int(*repoResp.Response.TotalCount) > 0 {
		log.Warningf("[tencent-tcr.PrepareForPush.createRepository.skip_exist] namespace=%s, repository=%s", namespace, repository)
		return
	}

	// 2. create
	var req = tcr.NewCreateRepositoryRequest()
	req.NamespaceName = &namespace
	req.RepositoryName = &repository
	req.RegistryId = a.registryID
	var resp = tcr.NewCreateRepositoryResponse()
	resp, err = a.tcrClient.CreateRepository(req)
	if err != nil {
		log.Debugf("[tencent-tcr.PrepareForPush.createRepository] error=%v", err)
		return
	}
	log.Debugf("[tencent-tcr.PrepareForPush.createRepository] resp=%#v", *resp)

	return
}

func (a *adapter) listNamespaces() (namespaces []string, err error) {
	if a.tcrClient == nil {
		err = errors.New("[tencent-tcr.listNamespaces] nil tcr client")
		return
	}

	// list namespaces
	var req = tcr.NewDescribeNamespacesRequest()
	req.RegistryId = a.registryID
	req.Limit = a.pageSize
	var resp = tcr.NewDescribeNamespacesResponse()

	var page int64
	for {
		req.Offset = &page
		resp, err = a.tcrClient.DescribeNamespaces(req)
		if err != nil {
			log.Debugf("[tencent-tcr.DescribeNamespaces] registryID=%s, error=%v", *a.registryID, err)
			return
		}

		for _, ns := range resp.Response.NamespaceList {
			namespaces = append(namespaces, *ns.Name)
		}

		if len(namespaces) >= int(*resp.Response.TotalCount) {
			break
		}
		page++
	}

	log.Debugf("[tencent-tcr.FetchArtifacts.listNamespaces] registryID=%s, namespaces[%d]=%s", *a.registryID, len(namespaces), namespaces)
	return
}

func (a *adapter) isNamespaceExist(namespace string) (exist bool, err error) {
	if a.tcrClient == nil {
		err = errors.New("[tencent-tcr.isNamespaceExist] nil tcr client")
		return
	}

	var req = tcr.NewDescribeNamespacesRequest()
	req.NamespaceName = &namespace
	req.RegistryId = a.registryID
	var resp = tcr.NewDescribeNamespacesResponse()
	resp, err = a.tcrClient.DescribeNamespaces(req)
	if err != nil {
		return
	}

	log.Warningf("[tencent-tcr.PrepareForPush.isNamespaceExist] namespace=%s, total=%d", namespace, *resp.Response.TotalCount)
	exist = isTcrNsExist(namespace, resp.Response.NamespaceList)

	return
}

func isTcrNsExist(name string, list []*tcr.TcrNamespaceInfo) (exist bool) {
	for _, ns := range list {
		if *ns.Name == name {
			exist = true
			return
		}
	}
	return
}

func (a *adapter) listReposByNamespace(namespace string) (repos []tcr.TcrRepositoryInfo, err error) {
	if a.tcrClient == nil {
		err = errors.New("[tencent-tcr.listReposByNamespace] nil tcr client")
		return
	}

	var req = tcr.NewDescribeRepositoriesRequest()
	req.RegistryId = a.registryID
	req.NamespaceName = common.StringPtr(namespace)
	req.Limit = a.pageSize
	var resp = tcr.NewDescribeRepositoriesResponse()

	var page int64 = 1
	var repositories []string
	for {
		req.Offset = common.Int64Ptr(page)
		resp, err = a.tcrClient.DescribeRepositories(req)
		if err != nil {
			log.Debugf("[tencent-tcr.listReposByNamespace.DescribeRepositories] registryID=%s, namespace=%s, error=%v", *a.registryID, namespace, err)
			return
		}

		size := len(resp.Response.RepositoryList)
		for i, repo := range resp.Response.RepositoryList {
			log.Debugf("[tencent-tcr.listReposByNamespace.DescribeRepositories] Retrives total=%d page=%d repo(%d/%d)=%s", *resp.Response.TotalCount, page, i, size, *repo.Name)
			repos = append(repos, *repo)
			repositories = append(repositories, *repo.Name)
		}

		log.Debugf("[tencent-tcr.listReposByNamespace.DescribeRepositories] total=%d now=%d page=%d,repositories=%v", *resp.Response.TotalCount, len(repos), page, repositories)
		if len(repos) == int(*resp.Response.TotalCount) {
			log.Debugf("[tencent-tcr.listReposByNamespace.DescribeRepositories] Retrives all repos.")
			break
		}
		page++
	}

	log.Debugf("[tencent-tcr.listReposByNamespace] registryID=%s, namespace=%s, repos=%d",
		*a.registryID, namespace, len(repos))
	return
}

func (a *adapter) getImages(namespace, repo, tag string) (images []*tcr.TcrImageInfo, imageNames []string, err error) {
	if a.tcrClient == nil {
		err = errors.New("[tencent-tcr.getImages] nil tcr client")
		return
	}

	if namespace != "" {
		repo = strings.Replace(repo, namespace, "", 1)
		repo = strings.Replace(repo, "/", "", 1)
	}

	var req = tcr.NewDescribeImagesRequest()
	req.RegistryId = a.registryID
	req.NamespaceName = &namespace
	req.RepositoryName = &repo
	req.Limit = a.pageSize
	if tag != "" {
		req.ImageVersion = &tag
	}
	var resp = tcr.NewDescribeImagesResponse()

	var page int64 = 1
	for {
		log.Debugf("[tencent-tcr.getImages] registryID=%s, namespace=%s, repo=%s, tag(s)=%d, page=%d",
			*a.registryID, namespace, repo, len(imageNames), page)
		req.Offset = &page
		resp, err = a.tcrClient.DescribeImages(req)
		if err != nil {
			log.Debugf("[tencent-tcr.getImages.DescribeImages] registryID=%s, namespace=%s, repo=%s, error=%v", *a.registryID, namespace, repo, err)
			return
		}

		images = resp.Response.ImageInfoList
		for _, image := range resp.Response.ImageInfoList {
			imageNames = append(imageNames, *image.ImageVersion)
		}

		if len(imageNames) == int(*resp.Response.TotalCount) {
			break
		}
		page++
	}

	log.Debugf("[tencent-tcr.getImages] registryID=%s, namespace=%s, repo=%s, tags[%d]=%v\n", *a.registryID, namespace, repo, len(imageNames), imageNames)
	return
}

func (a *adapter) deleteImage(namespace, repository, reference string) (err error) {
	var req = tcr.NewDeleteImageRequest()
	req.RegistryId = a.registryID
	req.NamespaceName = common.StringPtr(namespace)
	req.RepositoryName = common.StringPtr(repository)
	req.ImageVersion = common.StringPtr(reference)

	_, err = a.tcrClient.DeleteImage(req)
	if err != nil {
		log.Errorf("[tencent-tcr.deleteImage.DeleteImage] failed. namespace=%s, repository=%s, tag=%s, error=%s", namespace, repository, reference, err.Error())
	}

	return
}
