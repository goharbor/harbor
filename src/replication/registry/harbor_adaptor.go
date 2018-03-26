package registry

import (
	"github.com/vmware/harbor/src/common/dao"
	common_models "github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/replication"
	"github.com/vmware/harbor/src/replication/models"
	"github.com/vmware/harbor/src/ui/utils"
)

// TODO refacotor the methods of HarborAdaptor by caling Harbor's API

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
	repos, err := dao.GetRepositories(&common_models.RepositoryQuery{
		ProjectName: namespace,
	})
	if err != nil {
		log.Errorf("failed to get repositories under namespace %s: %v", namespace, err)
		return nil
	}

	repositories := []models.Repository{}
	for _, repo := range repos {
		repositories = append(repositories, models.Repository{
			Name: repo.Name,
		})
	}
	return repositories
}

//GetRepository is used to get the repository with the specified name under the specified namespace
func (ha *HarborAdaptor) GetRepository(name string, namespace string) models.Repository {
	return models.Repository{}
}

//GetTags is used to get all the tags of the specified repository under the namespace
func (ha *HarborAdaptor) GetTags(repositoryName string, namespace string) []models.Tag {
	client, err := utils.NewRepositoryClientForUI("harbor-ui", repositoryName)
	if err != nil {
		log.Errorf("failed to create registry client: %v", err)
		return nil
	}

	ts, err := client.ListTag()
	if err != nil {
		log.Errorf("failed to get tags of repository %s: %v", repositoryName, err)
		return nil
	}

	tags := []models.Tag{}
	for _, t := range ts {
		tags = append(tags, models.Tag{
			Name: t,
		})
	}

	return tags
}

//GetTag is used to get the tag with the specified name of the repository under the namespace
func (ha *HarborAdaptor) GetTag(name string, repositoryName string, namespace string) models.Tag {
	return models.Tag{}
}
