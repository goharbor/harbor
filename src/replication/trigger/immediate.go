package trigger

import (
	"github.com/vmware/harbor/src/replication"
)

//ImmediateTrigger will setup watcher at the image pushing action to fire
//replication event at pushing happening time.
type ImmediateTrigger struct {
	params ImmediateParam
}

//NewImmediateTrigger is constructor of ImmediateTrigger
func NewImmediateTrigger(params ImmediateParam) *ImmediateTrigger {
	return &ImmediateTrigger{
		params: params,
	}
}

//Kind is the implementation of same method defined in Trigger interface
func (st *ImmediateTrigger) Kind() string {
	return replication.TriggerKindImmediate
}

//Setup is the implementation of same method defined in Trigger interface
func (st *ImmediateTrigger) Setup() error {
	//TODO: Need more complicated logic here to handle partial updates
	for _, namespace := range st.params.Namespaces {
		wt := WatchItem{
			PolicyID:   st.params.PolicyID,
			Namespace:  namespace,
			OnDeletion: st.params.OnDeletion,
			OnPush:     true,
		}

		if err := DefaultWatchList.Add(wt); err != nil {
			return err
		}
	}
	return nil
}

//Unset is the implementation of same method defined in Trigger interface
func (st *ImmediateTrigger) Unset() error {
	return DefaultWatchList.Remove(st.params.PolicyID)
}
