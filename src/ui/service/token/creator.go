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

package token

import (
	"fmt"
	"github.com/docker/distribution/registry/auth/token"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
	"net/http"
	"strings"
)

var creatorMap map[string]TokenCreator

const (
	notary   = "harbor-notary"
	registry = "harbor-registry"
)

//InitCreators initialize the token creators for different services
func InitCreators() {
	creatorMap = make(map[string]TokenCreator)
	ext, err := config.ExtEndpoint()
	if err != nil {
		log.Warningf("Failed to get ext enpoint, err: %v, the token service will not be functional with notary requests", err)
	} else {
		creatorMap[notary] = &generalTokenCreator{
			validators: []ReqValidator{
				&basicAuthValidator{},
			},
			service: notary,
			filterMap: map[string]accessFilter{
				"repository": &repositoryFilter{
					parser: &endpointParser{
						endpoint: strings.Split(ext, "//")[1],
					},
				},
			},
		}
	}

	creatorMap[registry] = &generalTokenCreator{
		validators: []ReqValidator{
			&secretValidator{config.JobserviceSecret()},
			&basicAuthValidator{},
		},
		service: registry,
		filterMap: map[string]accessFilter{
			"repository": &repositoryFilter{
				//Workaround, had to use same service for both notary and registry
				parser: &endpointParser{
					endpoint: ext,
				},
			},
			"registry": &registryFilter{},
		},
	}
}

// TokenCreator creates a token ready to be served based on the http request.
type TokenCreator interface {
	create(r *http.Request) (*TokenJSON, error)
}

type imageParser interface {
	parse(s string) (*image, error)
}

type image struct {
	namespace string
	repo      string
	tag       string
}

type basicParser struct{}

func (b basicParser) parse(s string) (*image, error) {
	return parseImg(s)
}

type endpointParser struct {
	endpoint string
}

func (e endpointParser) parse(s string) (*image, error) {
	repo := strings.SplitN(s, "/", 2)
	if len(repo) < 2 {
		return nil, fmt.Errorf("Unable to parse image from string: %s", s)
	}
	//Workaround, need to use endpoint Parser to handle both cases.
	if strings.ContainsRune(repo[0], '.') {
		if repo[0] != e.endpoint {
			return nil, fmt.Errorf("Mismatch endpoint from string: %s, expected endpoint: %s", s, e.endpoint)
		}
		return parseImg(repo[1])
	} else {
		return parseImg(s)
	}
}

//build Image accepts a string like library/ubuntu:14.04 and build a image struct
func parseImg(s string) (*image, error) {
	repo := strings.SplitN(s, "/", 2)
	if len(repo) < 2 {
		return nil, fmt.Errorf("Unable to parse image from string: %s", s)
	}
	i := strings.SplitN(repo[1], ":", 2)
	res := &image{
		namespace: repo[0],
		repo:      i[0],
	}
	if len(i) == 2 {
		res.tag = i[1]
	}
	return res, nil
}

// An accessFilter will filter access based on userinfo
type accessFilter interface {
	filter(user userInfo, a *token.ResourceActions) error
}

type registryFilter struct {
}

func (reg registryFilter) filter(user userInfo, a *token.ResourceActions) error {
	//Do not filter if the request is to access registry catalog
	if a.Name != "catalog" {
		return fmt.Errorf("Unable to handle, type: %s, name: %s", a.Type, a.Name)
	}
	return nil
}

//repositoryFilter filters the access based on Harbor's permission model
type repositoryFilter struct {
	parser imageParser
}

func (rep repositoryFilter) filter(user userInfo, a *token.ResourceActions) error {
	//clear action list to assign to new acess element after perm check.
	a.Actions = []string{}
	img, err := rep.parser.parse(a.Name)
	if err != nil {
		return err
	}
	project := img.namespace
	permission := ""
	if user.allPerm {
		exist, err := dao.ProjectExists(project)
		if err != nil {
			log.Errorf("Error occurred in CheckExistProject: %v", err)
			//just leave empty permission
			return nil
		}
		if exist {
			permission = "RWM"
		} else {
			log.Infof("project %s does not exist, set empty permission for admin\n", project)
		}
	} else {
		permission, err = dao.GetPermission(user.name, project)
		if err != nil {
			log.Errorf("Error occurred in GetPermission: %v", err)
			//just leave empty permission
			return nil
		}
		if dao.IsProjectPublic(project) {
			permission += "R"
		}
	}
	a.Actions = permToActions(permission)
	return nil
}

type generalTokenCreator struct {
	validators []ReqValidator
	service    string
	filterMap  map[string]accessFilter
}

type unauthorizedError struct{}

func (e *unauthorizedError) Error() string {
	return "Unauthorized"
}

func (g generalTokenCreator) create(r *http.Request) (*TokenJSON, error) {
	var user *userInfo
	var err error
	var scopes []string
	scopeParm := r.URL.Query()["scope"]
	if len(scopeParm) > 0 {
		scopes = strings.Split(r.URL.Query()["scope"][0], " ")
	}
	log.Debugf("scopes: %v", scopes)
	for _, v := range g.validators {
		user, err = v.validate(r)
		if user != nil {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	if user == nil {
		if len(scopes) == 0 {
			return nil, &unauthorizedError{}
		}
		user = &userInfo{}
	}
	access := GetResourceActions(scopes)
	for _, a := range access {
		f, ok := g.filterMap[a.Type]
		if !ok {
			log.Warningf("No filter found for access type: %s, skip.", a.Type)
		}
		err = f.filter(*user, a)
		if err != nil {
			return nil, err
		}
	}
	return MakeToken(user.name, g.service, access)
}
