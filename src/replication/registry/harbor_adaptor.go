package registry

import (
	"github.com/vmware/harbor/src/replication"
	"github.com/vmware/harbor/src/replication/models"
)

//HarborAdaptor is defined to adapt the Harbor registry
type HarborAdaptor struct{}

//Kind returns the unique kind identifier of the adaptor
func (ha *HarborAdaptor) Kind() string {
	return replication.AdaptorKindHarbor
}

//GetNamespaces is ued to get all the namespaces
func (ha *HarborAdaptor) GetNamespaces() []models.Namespace {
	return nil
}

//GetNamespace is used to get the namespace with the specified name
func (ha *HarborAdaptor) GetNamespace(name string) models.Namespace {
	return models.Namespace{}
}

//GetRepositories is used to get all the repositories under the specified namespace
func (ha *HarborAdaptor) GetRepositories(namespace string) []models.Repository {
	return nil
}

//GetRepository is used to get the repository with the specified name under the specified namespace
func (ha *HarborAdaptor) GetRepository(name string, namespace string) models.Repository {
	return models.Repository{}
}

//GetTags is used to get all the tags of the specified repository under the namespace
func (ha *HarborAdaptor) GetTags(repositoryName string, namespace string) []models.Tag {
	return nil
}

//GetTag is used to get the tag with the specified name of the repository under the namespace
func (ha *HarborAdaptor) GetTag(name string, repositoryName string, namespace string) models.Tag {
	return models.Tag{}
}
