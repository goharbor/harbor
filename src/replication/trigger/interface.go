package trigger

//Interface is certian mechanism to know when fire the replication operation.
type Interface interface {
	//Kind indicates what type of the trigger is.
	Kind() string

	//Setup/enable the trigger; if failed, an error would be returned.
	Setup() error

	//Remove/disable the trigger; if failed, an error would be returned.
	Unset() error
}
