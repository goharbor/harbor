package registry

import (
	"github.com/vmware/harbor/src/replication/models"
)

//Adaptor defines the unified operations for all the supported registries such as Harbor or DockerHub.
//It's used to adapt the different interfaces provied by the different registry providers.
//Use external registry with restful api providing as example, these intrefaces may depends on the
//related restful apis like:
//  /api/vx/repositories/{namespace}/{repositoryName}/tags/{name}
//  /api/v0/accounts/{namespace}
type Adaptor interface {
	//Return the unique kind identifier of the adaptor
	Kind() string

	//Get all the namespaces
	GetNamespaces() []models.Namespace

	//Get the namespace with the specified name
	GetNamespace(name string) models.Namespace

	//Get all the repositories under the specified namespace
	GetRepositories(namespace string) []models.Repository

	//Get the repository with the specified name under the specified namespace
	GetRepository(name string, namespace string) models.Repository

	//Get all the tags of the specified repository under the namespace
	GetTags(repositoryName string, namespace string) []models.Tag

	//Get the tag with the specified name of the repository under the namespace
	GetTag(name string, repositoryName string, namespace string) models.Tag
}
