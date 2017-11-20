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
func (wl *WatchList) Remove() WatchItem {
	return WatchItem{}
}

//Update the watch item in the list
func (wl *WatchList) Update(updatedItem WatchItem) error {
	return nil
}

//Get the specified watch item
func (wl *WatchList) Get(namespace string) WatchItem {
	return WatchItem{}
}
