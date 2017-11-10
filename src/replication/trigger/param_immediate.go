package trigger

import (
	"errors"
)

//NOTES: Whether replicate the existing images when the type of trigger is
//'Immediate' is a once-effective setting which will not be persisted
// and kept as one parameter of 'Immediate' trigger. It will only be
//covered by the UI logic.

//ImmediateParam defines the parameter of immediate trigger
type ImmediateParam struct {
	//Basic parameters
	BasicParam

	//Namepace
	Namespace string
}

//Parse is the implementation of same method in TriggerParam interface
//NOTES: No need to implement this method for 'Immediate' trigger as
//it does not have any parameters with json format.
func (ip ImmediateParam) Parse(param string) error {
	return errors.New("Should NOT be called as it's not implemented")
}
