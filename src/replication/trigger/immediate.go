package trigger

import (
	"errors"

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
	if st.params.PolicyID <= 0 || len(st.params.Namespace) == 0 {
		return errors.New("Invalid parameters for Immediate trigger")
	}

	//TODO: Need more complicated logic here to handle partial updates
	wt := WatchItem{
		PolicyID:   st.params.PolicyID,
		Namespace:  st.params.Namespace,
		OnDeletion: st.params.OnDeletion,
		OnPush:     true,
	}

	return DefaultWatchList.Add(wt)
}

//Unset is the implementation of same method defined in Trigger interface
func (st *ImmediateTrigger) Unset() error {
	return errors.New("Not implemented")
}
