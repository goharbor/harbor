package models

//Namespace is the resource group/scope like project in Harbor and organization in docker hub.
type Namespace struct {
	//Name of the namespace
	Name string

	//Extensions to provide flexibility
	Metadata map[string]interface{}
}

//Repository is to keep the info of image repository.
type Repository struct {
	//Name of the repository
	Name string

	//Project reference of this repository belongs to
	Namespace Namespace

	//Extensions to provide flexibility
	Metadata map[string]interface{}
}

//Tag keeps the info of image with specified version
type Tag struct {
	//Name of the tag
	Name string

	//The repository reference of this tag belongs to
	Repository Repository

	//Extensions to provide flexibility
	Metadata map[string]interface{}
}
