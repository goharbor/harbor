package models

//FilterItem is the general data model represents the filtering resources which are used as input and output for the filters.
type FilterItem struct {

	//The kind of the filtering resources. Support 'project','repository' and 'tag' etc.
	Kind string

	//The key value of resource which can be used to filter out the resource matched with specified pattern.
	//E.g:
	//kind == 'project', value will be project name;
	//kind == 'repository', value will be repository name
	//kind == 'tag', value will be tag name.
	Value string

	//Extension placeholder.
	//To append more additional information if required by the filter.
	Metadata map[string]interface{}
}
