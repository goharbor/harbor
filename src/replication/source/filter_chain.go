package source

import (
	"github.com/vmware/harbor/src/replication/models"
)

//FilterChain is the interface to define the operations of coordinating multiple filters
//to work together as a whole pipeline.
//E.g:
//(original resources)---->[project filter]---->[repository filter]---->[tag filter]---->[......]---->(filter resources)
type FilterChain interface {
	//Build the filter chain with the filters provided;
	//if failed, an error will be returned.
	Build(filter []Filter) error

	//Return all the filters in the chain.
	Filters() []Filter

	//Filter the items and returned the filtered items via the appended filters in the chain.
	DoFilter(filterItems []models.FilterItem) []models.FilterItem
}
