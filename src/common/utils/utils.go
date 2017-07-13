// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"errors"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/harbor/src/common/utils/log"
)

// FormatEndpoint formats endpoint
func FormatEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimRight(endpoint, "/")
	if !strings.HasPrefix(endpoint, "http://") &&
		!strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	return endpoint
}

// ParseEndpoint parses endpoint to a URL
func ParseEndpoint(endpoint string) (*url.URL, error) {
	endpoint = FormatEndpoint(endpoint)

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return u, nil
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

// GenerateRandomString generates a random string
func GenerateRandomString() string {
	length := 32
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
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

// TestTCPConn tests TCP connection
// timeout: the total time before returning if something is wrong
// with the connection, in second
// interval: the interval time for retring after failure, in second
func TestTCPConn(addr string, timeout, interval int) error {
	success := make(chan int)
	cancel := make(chan int)

	go func() {
		for {
			select {
			case <-cancel:
				break
			default:
				conn, err := net.DialTimeout("tcp", addr, time.Duration(timeout)*time.Second)
				if err != nil {
					log.Errorf("failed to connect to tcp://%s, retry after %d seconds :%v",
						addr, interval, err)
					time.Sleep(time.Duration(interval) * time.Second)
					continue
				}
				conn.Close()
				success <- 1
				break
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

//ConvertMapToStruct is used to fill the specified struct with map.
func ConvertMapToStruct(object interface{}, valuesInMap map[string]interface{}) error {
	if object == nil {
		return fmt.Errorf("nil struct is not supported")
	}

	if reflect.TypeOf(object).Kind() != reflect.Ptr {
		return fmt.Errorf("object should be referred by pointer")
	}

	for k, v := range valuesInMap {
		if err := setField(object, k, v); err != nil {
			return err
		}
	}

	return nil
}

func setField(object interface{}, field string, value interface{}) error {
	structValue := reflect.ValueOf(object).Elem()

	structFieldValue := structValue.FieldByName(field)
	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", field)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set value for field %s", field)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return errors.New("Provided value type didn't match object field type")
	}

	structFieldValue.Set(val)

	return nil
}

// ParseProjectIDOrName parses value to ID(int64) or name(string)
func ParseProjectIDOrName(value interface{}) (int64, string, error) {
	if value == nil {
		return 0, "", errors.New("harborIDOrName is nil")
	}

	var id int64
	var name string
	switch value.(type) {
	case int:
		i := value.(int)
		id = int64(i)
		if id == 0 {
			return 0, "", fmt.Errorf("invalid ID: 0")
		}
	case int64:
		id = value.(int64)
		if id == 0 {
			return 0, "", fmt.Errorf("invalid ID: 0")
		}
	case string:
		name = value.(string)
		if len(name) == 0 {
			return 0, "", fmt.Errorf("empty name")
		}
	default:
		return 0, "", fmt.Errorf("unsupported type")
	}
	return id, name, nil
}
