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
/*
  [
    {
        "project_id": 1,
        "owner_id": 1,
        "name": "library",
        "creation_time": "2017-02-10T07:57:56Z",
        "creation_time_str": "",
        "deleted": 0,
        "owner_name": "",
        "public": 1,
        "togglable": true,
        "update_time": "2017-02-10T07:57:56Z",
        "current_user_role_id": 1,
        "repo_count": 0
    }
  ]
*/
export class Project {
    project_id: number;
    owner_id: number;
    name: string;
    creation_time: Date;
    creation_time_str: string;
    deleted: number;
    owner_name: string;
    togglable: boolean;
    update_time: Date;
    current_user_role_id: number;
    repo_count: number;
    chart_count: number;
    has_project_admin_role: boolean;
    is_member: boolean;
    role_name: string;
    metadata: {
      public: string | boolean;
      enable_content_trust: string | boolean;
      prevent_vul: string | boolean;
      severity: string;
      auto_scan: string | boolean;
      retention_id: number;
    };
    constructor () {
      this.metadata = <any>{};
      this.metadata.public = false;
    }
}
