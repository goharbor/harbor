// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	cronlib "github.com/robfig/cron/v3"

	"github.com/goharbor/harbor/src/lib/log"
)

// ParseEndpoint parses endpoint to a URL
func ParseEndpoint(endpoint string) (*url.URL, error) {
	endpoint = strings.Trim(endpoint, " ")
	endpoint = strings.TrimRight(endpoint, "/")
	if len(endpoint) == 0 {
		return nil, fmt.Errorf("empty URL")
	}
	i := strings.Index(endpoint, "://")
	if i >= 0 {
		scheme := endpoint[:i]
		if scheme != "http" && scheme != "https" {
			return nil, fmt.Errorf("invalid scheme: %s", scheme)
		}
	} else {
		endpoint = "http://" + endpoint
	}

	return url.ParseRequestURI(endpoint)
}

// ParseRepository splits a repository into two parts: project and rest
func ParseRepository(repository string) (project, rest string) {
	repository = strings.TrimLeft(repository, "/")
	repository = strings.TrimRight(repository, "/")
	if !strings.ContainsRune(repository, '/') {
		rest = repository
		return
	}
	index := strings.Index(repository, "/")
	project = repository[0:index]
	rest = repository[index+1:]
	return
}

// GenerateRandomStringWithLen generates a random string with length
func GenerateRandomStringWithLen(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	l := len(chars)
	result := make([]byte, length)
	_, err := rand.Read(result)
	if err != nil {
		log.Warningf("Error reading random bytes: %v", err)
	}
	for i := 0; i < length; i++ {
		result[i] = chars[int(result[i])%l]
	}
	return string(result)
}

// GenerateRandomString generate a random string with 32 byte length
func GenerateRandomString() string {
	return GenerateRandomStringWithLen(32)
}

// TestTCPConn tests TCP connection
// timeout: the total time before returning if something is wrong
// with the connection, in second
// interval: the interval time for retring after failure, in second
func TestTCPConn(addr string, timeout, interval int) error {
	success := make(chan int, 1)
	cancel := make(chan int, 1)

	go func() {
		n := 1

	loop:
		for {
			select {
			case <-cancel:
				break loop
			default:
				conn, err := net.DialTimeout("tcp", addr, time.Duration(n)*time.Second)
				if err != nil {
					log.Errorf("failed to connect to tcp://%s, retry after %d seconds :%v",
						addr, interval, err)
					n = n * 2
					time.Sleep(time.Duration(interval) * time.Second)
					continue
				}
				if err = conn.Close(); err != nil {
					log.Errorf("failed to close the connection: %v", err)
				}
				success <- 1
				break loop
			}
		}
	}()

	select {
	case <-success:
		return nil
	case <-time.After(time.Duration(timeout) * time.Second):
		cancel <- 1
		return fmt.Errorf("failed to connect to tcp:%s after %d seconds", addr, timeout)
	}
}

// ParseTimeStamp parse timestamp to time
func ParseTimeStamp(timestamp string) (*time.Time, error) {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, err
	}
	t := time.Unix(i, 0)
	return &t, nil
}

// ConvertMapToStruct is used to fill the specified struct with map.
func ConvertMapToStruct(object interface{}, values interface{}) error {
	if object == nil {
		return errors.New("nil struct is not supported")
	}

	if reflect.TypeOf(object).Kind() != reflect.Ptr {
		return errors.New("object should be referred by pointer")
	}

	bytes, err := json.Marshal(values)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, object)
}

// ParseProjectIDOrName parses value to ID(int64) or name(string)
func ParseProjectIDOrName(value interface{}) (int64, string, error) {
	if value == nil {
		return 0, "", errors.New("harborIDOrName is nil")
	}

	var id int64
	var name string
	switch v := value.(type) {
	case int, int64:
		id = reflect.ValueOf(v).Int()
	case string:
		name = value.(string)
	default:
		return 0, "", fmt.Errorf("unsupported type")
	}
	return id, name, nil
}

// SafeCastString -- cast a object to string saftely
func SafeCastString(value interface{}) string {
	if result, ok := value.(string); ok {
		return result
	}
	return ""
}

// SafeCastInt --
func SafeCastInt(value interface{}) int {
	if result, ok := value.(int); ok {
		return result
	}
	return 0
}

// SafeCastBool --
func SafeCastBool(value interface{}) bool {
	if result, ok := value.(bool); ok {
		return result
	}
	return false
}

// SafeCastFloat64 --
func SafeCastFloat64(value interface{}) float64 {
	if result, ok := value.(float64); ok {
		return result
	}
	return 0
}

// TrimLower ...
func TrimLower(str string) string {
	return strings.TrimSpace(strings.ToLower(str))
}

// GetStrValueOfAnyType return string format of any value, for map, need to convert to json
func GetStrValueOfAnyType(value interface{}) string {
	var strVal string
	if _, ok := value.(map[string]interface{}); ok {
		b, err := json.Marshal(value)
		if err != nil {
			log.Errorf("can not marshal json object, error %v", err)
			return ""
		}
		strVal = string(b)
	} else {
		switch val := value.(type) {
		case float64:
			strVal = strconv.FormatFloat(val, 'f', -1, 64)
		case float32:
			strVal = strconv.FormatFloat(float64(val), 'f', -1, 32)
		default:
			strVal = fmt.Sprintf("%v", value)
		}
	}
	return strVal
}

// IsIllegalLength ...
func IsIllegalLength(s string, min int, max int) bool {
	if min == -1 {
		return (len(s) > max)
	}
	if max == -1 {
		return (len(s) <= min)
	}
	return (len(s) < min || len(s) > max)
}

// ParseJSONInt ...
func ParseJSONInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	default:
		return 0, false
	}
}

// FindNamedMatches returns named matches of the regexp groups
func FindNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)

	results := map[string]string{}
	for i, name := range match {
		results[regex.SubexpNames()[i]] = name
	}
	return results
}

// NextSchedule return next scheduled time with a cron string and current time provided
// the cron string could contain 6 tokens
// if the cron string is invalid, it returns a zero time
func NextSchedule(cron string, curTime time.Time) time.Time {
	if len(cron) == 0 {
		return time.Time{}
	}
	cr := strings.TrimSpace(cron)
	s, err := CronParser().Parse(cr)
	if err != nil {
		log.Debugf("the cron string %v is invalid, error: %v", cron, err)
		return time.Time{}
	}
	return s.Next(curTime)
}

// CronParser returns the parser of cron string with format of "* * * * * *"
func CronParser() cronlib.Parser {
	return cronlib.NewParser(cronlib.Second | cronlib.Minute | cronlib.Hour | cronlib.Dom | cronlib.Month | cronlib.Dow)
}

// ValidateCronString check whether it is a valid cron string and whether the 1st field (indicating Seconds of time) of the cron string is a fixed value of 0 or not
func ValidateCronString(cron string) error {
	if len(cron) == 0 {
		return fmt.Errorf("empty cron string is invalid")
	}
	_, err := CronParser().Parse(cron)
	if err != nil {
		return err
	}
	cronParts := strings.Split(cron, " ")
	if len(cronParts) == 6 && cronParts[0] != "0" {
		return fmt.Errorf("the 1st field (indicating Seconds of time) of the cron setting must be 0")
	}
	return nil
}

// MostMatchSorter is a sorter for the most match, usually invoked in sort Less function
// usage:
//
//	sort.Slice(input, func(i, j int) bool {
//		return MostMatchSorter(input[i].GroupName, input[j].GroupName, matchWord)
//	})
// a is the field to be used for sorting, b is the other field, matchWord is the word to be matched
// the return value is true if a is less than b
// for example, search with "user",  input is {"harbor_user", "user", "users, "admin_user"}
// it returns with this order {"user", "users", "admin_user", "harbor_user"}

func MostMatchSorter(a, b string, matchWord string) bool {
	// exact match always first
	if a == matchWord {
		return true
	}
	if b == matchWord {
		return false
	}
	// sort by length, then sort by alphabet
	if len(a) == len(b) {
		return a < b
	}
	return len(a) < len(b)
}

// IsLocalPath checks if path is local
func IsLocalPath(path string) bool {
	return strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "//")
}
