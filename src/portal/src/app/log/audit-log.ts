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
 {
    "log_id": 3,
    "user_id": 0,
    "project_id": 0,
    "repo_name": "library/mysql",
    "repo_tag": "5.6",
    "guid": "",
    "operation": "push",
    "op_time": "2017-02-14T09:22:58Z",
    "username": "admin",
    "keywords": "",
    "BeginTime": "0001-01-01T00:00:00Z",
    "begin_timestamp": 0,
    "EndTime": "0001-01-01T00:00:00Z",
    "end_timestamp": 0
  }
*/
export class AuditLog {
  log_id: number | string;
  project_id: number | string;
  username: string;
  repo_name: string;
  repo_tag: string;
  operation: string;
  op_time: Date;
  begin_timestamp: number | string;
  end_timestamp: number | string;
  keywords: string;
  page: number | string;
  page_size: number | string;
  fromTime: string;
  toTime: string;
}
