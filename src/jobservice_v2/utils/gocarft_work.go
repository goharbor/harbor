package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gocraft/work"
)

//Functions defined here are mainly from dep lib "github.com/gocraft/work".
//Only for compatible

//MakeUniquePeriodicID creates id for the periodic job.
func MakeUniquePeriodicID(name, spec string, epoch int64) string {
	return fmt.Sprintf("periodic:job:%s:%s:%d", name, spec, epoch)
}

//RedisNamespacePrefix ... Same with 'KeyNamespacePrefix', only for compatiblity.
func RedisNamespacePrefix(namespace string) string {
	return KeyNamespacePrefix(namespace)
}

//RedisKeyScheduled returns key of scheduled job.
func RedisKeyScheduled(namespace string) string {
	return RedisNamespacePrefix(namespace) + "scheduled"
}

//RedisKeyLastPeriodicEnqueue returns key of timestamp if last periodic enqueue.
func RedisKeyLastPeriodicEnqueue(namespace string) string {
	return RedisNamespacePrefix(namespace) + "last_periodic_enqueue"
}

var nowMock int64

//NowEpochSeconds ...
func NowEpochSeconds() int64 {
	if nowMock != 0 {
		return nowMock
	}
	return time.Now().Unix()
}

//SetNowEpochSecondsMock ...
func SetNowEpochSecondsMock(t int64) {
	nowMock = t
}

//SerializeJob encodes work.Job to json data.
func SerializeJob(job *work.Job) ([]byte, error) {
	return json.Marshal(job)
}

//DeSerializeJob decodes bytes to ptr of work.Job.
func DeSerializeJob(jobBytes []byte) (*work.Job, error) {
	var j work.Job
	err := json.Unmarshal(jobBytes, &j)

	return &j, err
}
