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
/*
  {
    "id": 1,
    "project_id": 1,
    "project_name": "library",
    "target_id": 1,
    "target_name": "target_01",
    "name": "sync_01",
    "enabled": 0,
    "description": "sync_01 desc.",
    "cron_str": "",
    "start_time": "0001-01-01T00:00:00Z",
    "creation_time": "2017-02-24T06:41:52Z",
    "update_time": "2017-02-24T06:41:52Z",
    "error_job_count": 0,
    "deleted": 0
  }
*/

export class Policy {
  id: number;
  project_id: number;
  project_name: string;
  target_id: number;
  target_name: string;
  name: string;
  enabled: number;
  description: string;
  cron_str: string;
  start_time: Date;
  creation_time: Date;
  update_time: Date;
  error_job_count: number;
  deleted: number;
}