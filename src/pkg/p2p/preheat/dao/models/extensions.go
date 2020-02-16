package models

import (
	"errors"
	"fmt"
	"strings"
)

// ImageRepository represents the image repository name
// e.g: library/ubuntu:latest
type ImageRepository string

// Valid checks if the repository name is valid
func (ir ImageRepository) Valid() bool {
	if len(ir) == 0 {
		return false
	}

	trimName := strings.TrimSpace(string(ir))
	segments := strings.SplitN(trimName, "/", 2)
	if len(segments) != 2 {
		return false
	}

	nameAndTag := segments[1]
	subSegments := strings.SplitN(nameAndTag, ":", 2)
	if len(subSegments) != 2 {
		return false
	}

	return true
}

// Name returns the name of the image repository
func (ir ImageRepository) Name() string {
	// No check here, should call Valid() before calling name
	segments := strings.SplitN(string(ir), ":", 2)
	if len(segments) == 0 {
		return ""
	}

	return segments[0]
}

// Tag returns the tag of the image repository
func (ir ImageRepository) Tag() string {
	// No check here, should call Valid() before calling name
	segments := strings.SplitN(string(ir), ":", 2)
	if len(segments) < 2 {
		return ""
	}

	return segments[1]
}

// changableProperties contains names of the changable properties
type changableProperties []string

// Append changable property
func (cp *changableProperties) Append(prop ...string) *changableProperties {
	*cp = append(*cp, prop...)
	return cp
}

// Match confirms if the provided prop are changable
func (cp *changableProperties) Match(prop string) bool {
	for _, p := range *cp {
		if p == prop {
			return true
		}
	}

	return false
}

// TrackStatus indicates the status info
type TrackStatus string

// Valid the status
func (ts TrackStatus) Valid() bool {
	switch string(ts) {
	case PreheatingStatusPending,
		PreheatingStatusRunning,
		PreheatingStatusSuccess,
		PreheatingStatusFail:
		return true
	default:
	}

	return false
}

// Success status
func (ts TrackStatus) Success() bool {
	return string(ts) == PreheatingStatusSuccess
}

// Running status
func (ts TrackStatus) Running() bool {
	return string(ts) == PreheatingStatusRunning
}

// Fail status
func (ts TrackStatus) Fail() bool {
	return string(ts) == PreheatingStatusFail
}

// Pending status
func (ts TrackStatus) Pending() bool {
	return string(ts) == PreheatingStatusPending
}

// String value of status
func (ts TrackStatus) String() string {
	return string(ts)
}

// For querying
var theChangableProperties = initChangableProperties()

// PropertySet for incremental updating
type PropertySet map[string]interface{}

// Apply the properties
func (ps PropertySet) Apply(metadata *Metadata) error {
	if metadata == nil {
		return errors.New("nil metadata")
	}

	for k, v := range ps {
		if !theChangableProperties.Match(k) {
			return fmt.Errorf("property '%s' is not changable", k)
		}

		pv := PropertyValue{v}
		if err := ps.assign(k, pv, metadata); err != nil {
			return err
		}
	}

	return nil
}

func (ps PropertySet) assign(prop string, val PropertyValue, metadata *Metadata) error {
	switch prop {
	case "endpoint":
		if v := val.String(); len(v) > 0 {
			metadata.Endpoint = v
			return nil
		}
	case "enabled":
		v := val.Bool()
		metadata.Enabled = v // invalid value will be treated as false
		return nil
	case "auth_mode":
		if v := val.String(); len(v) > 0 {
			metadata.AuthMode = v
			return nil
		}
	case "auth_data":
		if v := val.StringMap(); v != nil {
			metadata.AuthData = v
			return nil
		}
	case "description":
		if v := val.String(); len(v) > 0 {
			metadata.Description = v
			return nil
		}
	default:
	}

	return fmt.Errorf("assign instance property error %s: %v", prop, val)
}

// PropertyValue for keeping value of property
type PropertyValue struct {
	val interface{}
}

// String value
func (pv PropertyValue) String() string {
	if pv.val == nil {
		return ""
	}

	v, ok := pv.val.(string)
	if ok {
		return v
	}

	return ""
}

// Bool value
func (pv PropertyValue) Bool() bool {
	if pv.val == nil {
		return false
	}

	v, ok := pv.val.(bool)
	if ok {
		return v
	}

	return false
}

// StringMap value
func (pv PropertyValue) StringMap() map[string]string {
	if pv.val == nil {
		return nil
	}

	v, ok := pv.val.(map[string]interface{})
	if ok {
		stringMap := map[string]string{}
		for k, v := range v {
			stringMap[k] = fmt.Sprintf("%s", v)
		}

		return stringMap
	}

	return nil
}

// initChangableProperties is used as a initializer of ChangableProperties
func initChangableProperties() *changableProperties {
	cp := make(changableProperties, 0)
	cpr := &cp
	cpr.Append("auth_mode").
		Append("auth_data").
		Append("endpoint").
		Append("enabled").
		Append("description")

	return cpr
}
