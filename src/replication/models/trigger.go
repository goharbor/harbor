package models

import (
	"fmt"

	"github.com/astaxie/beego/validation"
	"github.com/vmware/harbor/src/replication"
)

//Trigger is replication launching approach definition
type Trigger struct {
	//The name of the trigger
	Kind string `json:"kind"`

	//The parameters with json text format required by the trigger
	Param string `json:"param"`
}

// Valid ...
func (t *Trigger) Valid(v *validation.Validation) {
	if !(t.Kind == replication.TriggerKindImmediate ||
		t.Kind == replication.TriggerKindManual ||
		t.Kind == replication.TriggerKindSchedule) {
		v.SetError("kind", fmt.Sprintf("invalid trigger kind: %s", t.Kind))
	}
}
