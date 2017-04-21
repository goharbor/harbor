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
    "id": "2",
    "name": "library/mysql",
    "owner_id": 1,
    "project_id": 1,
    "description": "",
    "pull_count": 0,
    "star_count": 0,
    "tags_count": 1,
    "creation_time": "2017-02-14T09:22:58Z",
    "update_time": "0001-01-01T00:00:00Z"
  }
*/

export class Repository {
  id: number;
  name: string;
  owner_id: number;
  project_id: number;
  description: string;
  pull_count: number;
  start_count: number;
  tags_count: number;
  creation_time: Date;
  update_time: Date;

  constructor(name: string, tags_count: number) {
    this.name = name;
    this.tags_count = tags_count;
  }
}