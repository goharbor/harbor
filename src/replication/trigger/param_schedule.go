package trigger

import (
	"encoding/json"
	"errors"
)

//ScheduleParam defines the parameter of schedule trigger
type ScheduleParam struct {
	//Basic parameters
	BasicParam

	//Daily or weekly
	Type string

	//Optional, only used when type is 'weekly'
	Weekday int8

	//The time offset with the UTC 00:00 in seconds
	Offtime int64
}

//Parse is the implementation of same method in TriggerParam interface
func (stp ScheduleParam) Parse(param string) error {
	if len(param) == 0 {
		return errors.New("Parameter of schedule trigger should not be empty")
	}

	return json.Unmarshal([]byte(param), &stp)
}
