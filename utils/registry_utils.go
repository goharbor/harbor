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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/astaxie/beego"
)

const sessionCookie = "beegosessionID"

func BuildRegistryURL(segments ...string) string {
	registryURL := os.Getenv("REGISTRY_URL")
	if registryURL == "" {
		registryURL = "http://localhost:5000"
	}
	url := registryURL + "/v2"
	for _, s := range segments {
		if s == "v2" {
			beego.Error("Unnecessary v2 in", segments)
			continue
		}
		url += "/" + s
	}
	return url
}

func HTTPGet(URL, sessionID, username, password string) ([]byte, error) {
	response, err := http.Get(URL)
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
		str := strings.Split(authenticate, " ")[1]
		beego.Trace("url: " + URL)
		beego.Trace("Authentication Header: " + str)
		var realm string
		var service string
		var scope string
		strs := strings.Split(str, ",")
		for _, s := range strs {
			if strings.Contains(s, "realm") {
				realm = s
			} else if strings.Contains(s, "service") {
				service = s
			} else if strings.Contains(s, "scope") {
				strings.HasSuffix(URL, "v2/_catalog")
				scope = s
			}
		}
		realm = strings.Split(realm, "\"")[1]
		service = strings.Split(service, "\"")[1]
		scope = strings.Split(scope, "\"")[1]

		authURL := realm + "?service=" + service + "&scope=" + scope
		//skip certificate check if token service is https.
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		request, err := http.NewRequest("GET", authURL, nil)
		if err != nil {
			return nil, err
		}
		if len(sessionID) > 0 {
			cookie := &http.Cookie{Name: sessionCookie, Value: sessionID, Path: "/"}
			request.AddCookie(cookie)
		} else {
			request.SetBasicAuth(username, password)
		}
		response, err = client.Do(request)
		if err != nil {
			return nil, err
		}
		result, err = ioutil.ReadAll(response.Body)

		defer response.Body.Close()
		if err != nil {
			return nil, err
		}
		if response.StatusCode == http.StatusOK {
			tt := make(map[string]string)
			json.Unmarshal(result, &tt)
			request, err = http.NewRequest("GET", URL, nil)
			if err != nil {
				return nil, err
			}
			request.Header.Add("Authorization", "Bearer "+tt["token"])
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
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
			result, err = ioutil.ReadAll(response.Body)
			if err != nil {
				return nil, err
			}
			defer response.Body.Close()

			return result, nil
		}
		return nil, errors.New(string(result))
	} else {
		return nil, errors.New(string(result))
	}
}
