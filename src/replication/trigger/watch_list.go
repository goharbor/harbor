package trigger

//DefaultWatchList is the default instance of WatchList
var DefaultWatchList = &WatchList{}

//WatchList contains the items which should be evaluated for replication
//when image pushing or deleting happens.
type WatchList struct{}

//WatchItem keeps the related data for evaluation in WatchList.
type WatchItem struct {
	//ID of policy
	PolicyID int64

	//Corresponding namespace
	Namespace string

	//For deletion event
	OnDeletion bool

	//For pushing event
	OnPush bool
}

//Add item to the list and persist into DB
func (wl *WatchList) Add(item WatchItem) error {
	return nil
}

//Remove the specified watch item from list
func (wl *WatchList) Remove(policyID int64) error {
	return nil
}

//Get the watch items according to the namespace and operation
func (wl *WatchList) Get(namespace, operation string) ([]WatchItem, error) {
	return []WatchItem{}, nil
}
