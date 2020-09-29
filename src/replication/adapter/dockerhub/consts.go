package dockerhub

import "fmt"

const (
	baseURL             = "https://hub.docker.com"
	registryURL         = "https://registry-1.docker.io"
	loginPath           = "/v2/users/login/"
	listNamespacePath   = "/v2/repositories/namespaces"
	createNamespacePath = "/v2/orgs/"

	metadataKeyCompany  = "company"
	metadataKeyFullName = "fullName"
)

func getNamespacePath(namespace string) string {
	return fmt.Sprintf("/v2/orgs/%s/", namespace)
}

func listReposPath(namespace, name string, page, pageSize int) string {
	if len(name) == 0 {
		return fmt.Sprintf("/v2/repositories/%s/?page=%d&page_size=%d", namespace, page, pageSize)
	}

	return fmt.Sprintf("/v2/repositories/%s/?name=%s&page=%d&page_size=%d", namespace, name, page, pageSize)
}

func listTagsPath(namespace, repo string, page, pageSize int) string {
	return fmt.Sprintf("/v2/repositories/%s/%s/tags/?page=%d&page_size=%d", namespace, repo, page, pageSize)
}

func deleteTagPath(namespace, repo, tag string) string {
	return fmt.Sprintf("/v2/repositories/%s/%s/tags/%s/", namespace, repo, tag)
}
