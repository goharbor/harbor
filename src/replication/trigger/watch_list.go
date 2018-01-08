package trigger

import (
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
)

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
	_, err := dao.DefaultDatabaseWatchItemDAO.Add(
		&models.WatchItem{
			PolicyID:   item.PolicyID,
			Namespace:  item.Namespace,
			OnPush:     item.OnPush,
			OnDeletion: item.OnDeletion,
		})
	return err
}

//Remove the specified watch item from list
func (wl *WatchList) Remove(policyID int64) error {
	return dao.DefaultDatabaseWatchItemDAO.DeleteByPolicyID(policyID)
}

//Get the watch items according to the namespace and operation
func (wl *WatchList) Get(namespace, operation string) ([]WatchItem, error) {
	items, err := dao.DefaultDatabaseWatchItemDAO.Get(namespace, operation)
	if err != nil {
		return nil, err
	}

	watchItems := []WatchItem{}
	for _, item := range items {
		watchItems = append(watchItems, WatchItem{
			PolicyID:   item.PolicyID,
			Namespace:  item.Namespace,
			OnPush:     item.OnPush,
			OnDeletion: item.OnDeletion,
		})
	}

	return watchItems, nil
}
