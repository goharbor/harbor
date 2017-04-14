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
export class AppConfig {
    constructor(){
        //Set default value
        this.with_notary = false;
        this.with_admiral = false;
        this.admiral_endpoint = "";
        this.auth_mode = "db_auth";
        this.registry_url = "";
        this.project_creation_restriction = "everyone";
        this.self_registration = true;
        this.has_ca_root = false;
        this.harbor_version = "0.5.0";//default
    }
    
    with_notary: boolean;
    with_admiral: boolean;
    admiral_endpoint: string;
    auth_mode: string;
    registry_url: string;
    project_creation_restriction: string;
    self_registration: boolean;
    has_ca_root: boolean;
    harbor_version: string;
}