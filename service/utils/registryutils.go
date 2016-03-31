/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/vmware/harbor/utils/log"
)

// BuildRegistryURL ...
func BuildRegistryURL(segments ...string) string {
	registryURL := os.Getenv("REGISTRY_URL")
	if registryURL == "" {
		registryURL = "http://localhost:5000"
	}
	url := registryURL + "/v2"
	for _, s := range segments {
		if s == "v2" {
			log.Debugf("unnecessary v2 in %v", segments)
			continue
		}
		url += "/" + s
	}
	return url
}

// RegistryAPIGet triggers GET request to the URL which is the endpoint of registry and returns the response body.
// It will attach a valid jwt token to the request if registry requires.
func RegistryAPIGet(url, username string) ([]byte, error) {

	log.Debugf("Registry API url: %s", url)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	result, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		return result, nil
	} else if response.StatusCode == http.StatusUnauthorized {
		authenticate := response.Header.Get("WWW-Authenticate")
		log.Debugf("authenticate header: %s", authenticate)
		var service string
		var scope string
		re := regexp.MustCompile(`service=\"(.*?)\".*scope=\"(.*?)\"`)
		res := re.FindStringSubmatch(authenticate)
		if len(res) > 2 {
			service = res[1]
			scope = res[2]
		}
		token, err := GenTokenForUI(username, service, scope)
		if err != nil {
			return nil, err
		}
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Add("Authorization", "Bearer "+token)
		client := &http.Client{}
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			//	log.Infof("via length: %d\n", len(via))
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			for k, v := range via[0].Header {
				if _, ok := req.Header[k]; !ok {
					req.Header[k] = v
				}
			}
			return nil
		}
		response, err = client.Do(request)
		if err != nil {
			return nil, err
		}
		if response.StatusCode != http.StatusOK {
			errMsg := fmt.Sprintf("Unexpected return code from registry: %d", response.StatusCode)
			log.Error(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		result, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		return result, nil
	} else {
		return nil, errors.New(string(result))
	}
}
