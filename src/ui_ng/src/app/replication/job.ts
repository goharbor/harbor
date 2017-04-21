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
    "status": "running",
    "repository": "library/mysql",
    "policy_id": 1,
    "operation": "transfer",
    "tags": null,
    "creation_time": "2017-02-24T06:44:04Z",
    "update_time": "2017-02-24T06:44:04Z"
  }

*/
export class Job {
  id: number;
  status: string;
  repository: string;
  policy_id: number;
  operation: string;
  tags: string;
  creation_time: Date;
  update_time: Date;
}