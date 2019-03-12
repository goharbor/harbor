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
