package models

import (
	"fmt"

	"github.com/astaxie/beego/validation"
	"github.com/vmware/harbor/src/replication"
)

//FilterItem is the general data model represents the filtering resources which are used as input and output for the filters.
type FilterItem struct {

	//The kind of the filtering resources. Support 'project','repository' and 'tag' etc.
	Kind string `json:"kind"`

	//The key value of resource which can be used to filter out the resource matched with specified pattern.
	//E.g:
	//kind == 'project', value will be project name;
	//kind == 'repository', value will be repository name
	//kind == 'tag', value will be tag name.
	Value string `json:"value"`

	Operation string `json:"operation"`

	//Extension placeholder.
	//To append more additional information if required by the filter.
	Metadata map[string]interface{} `json:"metadata"`
}

// Valid ...
func (f *FilterItem) Valid(v *validation.Validation) {
	if !(f.Kind == replication.FilterItemKindProject ||
		f.Kind == replication.FilterItemKindRepository ||
		f.Kind == replication.FilterItemKindTag) {
		v.SetError("kind", fmt.Sprintf("invalid filter kind: %s", f.Kind))
	}

	if len(f.Value) == 0 {
		v.SetError("value", "filter value can not be empty")
	}
}
