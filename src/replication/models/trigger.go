package models

//Trigger is replication launching approach definition
type Trigger struct {
	//The name of the trigger
	Name string

	//The parameters with json text format required by the trigger
	Param string
}
