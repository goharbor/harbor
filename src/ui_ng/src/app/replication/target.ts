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
    "endpoint": "http://10.117.4.151",
    "name": "target_01",
    "username": "admin",
    "password": "Harbor12345",
    "type": 0,
    "creation_time": "2017-02-24T06:41:52Z",
    "update_time": "2017-02-24T06:41:52Z"
  }
*/

export class Target {
  id: number;
  endpoint: string;
  name: string;
  username: string;
  password: string;
  type: number;
  creation_time: Date;
  update_time: Date;
}