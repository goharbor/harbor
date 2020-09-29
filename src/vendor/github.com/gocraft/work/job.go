package work

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
)

// Job represents a job.
type Job struct {
	// Inputs when making a new job
	Name       string                 `json:"name,omitempty"`
	ID         string                 `json:"id"`
	EnqueuedAt int64                  `json:"t"`
	Args       map[string]interface{} `json:"args"`
	Unique     bool                   `json:"unique,omitempty"`

	// Inputs when retrying
	Fails    int64  `json:"fails,omitempty"` // number of times this job has failed
	LastErr  string `json:"err,omitempty"`
	FailedAt int64  `json:"failed_at,omitempty"`

	rawJSON      []byte
	dequeuedFrom []byte
	inProgQueue  []byte
	argError     error
	observer     *observer
}

// Q is a shortcut to easily specify arguments for jobs when enqueueing them.
// Example: e.Enqueue("send_email", work.Q{"addr": "test@example.com", "track": true})
type Q map[string]interface{}

func newJob(rawJSON, dequeuedFrom, inProgQueue []byte) (*Job, error) {
	var job Job
	err := json.Unmarshal(rawJSON, &job)
	if err != nil {
		return nil, err
	}
	job.rawJSON = rawJSON
	job.dequeuedFrom = dequeuedFrom
	job.inProgQueue = inProgQueue
	return &job, nil
}

func (j *Job) serialize() ([]byte, error) {
	return json.Marshal(j)
}

// setArg sets a single named argument on the job.
func (j *Job) setArg(key string, val interface{}) {
	if j.Args == nil {
		j.Args = make(map[string]interface{})
	}
	j.Args[key] = val
}

func (j *Job) failed(err error) {
	j.Fails++
	j.LastErr = err.Error()
	j.FailedAt = nowEpochSeconds()
}

// Checkin will update the status of the executing job to the specified messages. This message is visible within the web UI. This is useful for indicating some sort of progress on very long running jobs. For instance, on a job that has to process a million records over the course of an hour, the job could call Checkin with the current job number every 10k jobs.
func (j *Job) Checkin(msg string) {
	if j.observer != nil {
		j.observer.observeCheckin(j.Name, j.ID, msg)
	}
}

// ArgString returns j.Args[key] typed to a string. If the key is missing or of the wrong type, it sets an argument error
// on the job. This function is meant to be used in the body of a job handling function while extracting arguments,
// followed by a single call to j.ArgError().
func (j *Job) ArgString(key string) string {
	v, ok := j.Args[key]
	if ok {
		typedV, ok := v.(string)
		if ok {
			return typedV
		}
		j.argError = typecastError("string", key, v)
	} else {
		j.argError = missingKeyError("string", key)
	}
	return ""
}

// ArgInt64 returns j.Args[key] typed to an int64. If the key is missing or of the wrong type, it sets an argument error
// on the job. This function is meant to be used in the body of a job handling function while extracting arguments,
// followed by a single call to j.ArgError().
func (j *Job) ArgInt64(key string) int64 {
	v, ok := j.Args[key]
	if ok {
		rVal := reflect.ValueOf(v)
		if isIntKind(rVal) {
			return rVal.Int()
		} else if isUintKind(rVal) {
			vUint := rVal.Uint()
			if vUint <= math.MaxInt64 {
				return int64(vUint)
			}
		} else if isFloatKind(rVal) {
			vFloat64 := rVal.Float()
			vInt64 := int64(vFloat64)
			if vFloat64 == math.Trunc(vFloat64) && vInt64 <= 9007199254740892 && vInt64 >= -9007199254740892 {
				return vInt64
			}
		}
		j.argError = typecastError("int64", key, v)
	} else {
		j.argError = missingKeyError("int64", key)
	}
	return 0
}

// ArgFloat64 returns j.Args[key] typed to a float64. If the key is missing or of the wrong type, it sets an argument error
// on the job. This function is meant to be used in the body of a job handling function while extracting arguments,
// followed by a single call to j.ArgError().
func (j *Job) ArgFloat64(key string) float64 {
	v, ok := j.Args[key]
	if ok {
		rVal := reflect.ValueOf(v)
		if isIntKind(rVal) {
			return float64(rVal.Int())
		} else if isUintKind(rVal) {
			return float64(rVal.Uint())
		} else if isFloatKind(rVal) {
			return rVal.Float()
		}
		j.argError = typecastError("float64", key, v)
	} else {
		j.argError = missingKeyError("float64", key)
	}
	return 0.0
}

// ArgBool returns j.Args[key] typed to a bool. If the key is missing or of the wrong type, it sets an argument error
// on the job. This function is meant to be used in the body of a job handling function while extracting arguments,
// followed by a single call to j.ArgError().
func (j *Job) ArgBool(key string) bool {
	v, ok := j.Args[key]
	if ok {
		typedV, ok := v.(bool)
		if ok {
			return typedV
		}
		j.argError = typecastError("bool", key, v)
	} else {
		j.argError = missingKeyError("bool", key)
	}
	return false
}

// ArgError returns the last error generated when extracting typed params. Returns nil if extracting the args went fine.
func (j *Job) ArgError() error {
	return j.argError
}

func isIntKind(v reflect.Value) bool {
	k := v.Kind()
	return k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64
}

func isUintKind(v reflect.Value) bool {
	k := v.Kind()
	return k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64
}

func isFloatKind(v reflect.Value) bool {
	k := v.Kind()
	return k == reflect.Float32 || k == reflect.Float64
}

func missingKeyError(jsonType, key string) error {
	return fmt.Errorf("looking for a %s in job.Arg[%s] but key wasn't found", jsonType, key)
}

func typecastError(jsonType, key string, v interface{}) error {
	actualType := reflect.TypeOf(v)
	return fmt.Errorf("looking for a %s in job.Arg[%s] but value wasn't right type: %v(%v)", jsonType, key, actualType, v)
}
