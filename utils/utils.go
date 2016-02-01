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
	"encoding/base64"
	"strings"

	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego"
)

type Repository struct {
	Name string
}

func ParseBasicAuth(authorization []string) (username, password string) {
	if authorization == nil || len(authorization) == 0 {
		beego.Debug("Authorization header is not set.")
		return "", ""
	}
	auth := strings.SplitN(authorization[0], " ", 2)
	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)
	return pair[0], pair[1]
}

func (r *Repository) GetProject() string {
	if !strings.ContainsRune(r.Name, '/') {
		return ""
	}
	return r.Name[0:strings.LastIndex(r.Name, "/")]
}

type ProjectSorter struct {
	Projects []models.Project
}

func (ps *ProjectSorter) Len() int {
	return len(ps.Projects)
}

func (ps *ProjectSorter) Less(i, j int) bool {
	return ps.Projects[i].Name < ps.Projects[j].Name
}

func (ps *ProjectSorter) Swap(i, j int) {
	ps.Projects[i], ps.Projects[j] = ps.Projects[j], ps.Projects[i]
}
