package source

import (
	"github.com/vmware/harbor/src/replication/models"
)

//Filter define the operations of selecting the matched resources from the candidates
//according to the specified pattern.
type Filter interface {
	//Initialize the filter
	Init() error

	//Return the convertor if existing or nil if never set
	GetConvertor() Convertor

	//Filter the items
	DoFilter(filterItems []models.FilterItem) []models.FilterItem
}
