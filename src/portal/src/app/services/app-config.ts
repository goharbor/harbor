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

import { ClairDBStatus } from '../shared/services';

export class AppConfig {
    with_trivy: boolean;
    admiral_endpoint: string;
    auth_mode: string;
    primary_auth_mode: boolean;
    registry_url: string;
    project_creation_restriction: string;
    self_registration: boolean;
    has_ca_root: boolean;
    harbor_version: string;
    clair_vulnerability_status?: ClairDBStatus;
    next_scan_all: number;
    registry_storage_provider_name: string;
    read_only: boolean;
    show_popular_repo: boolean;
    banner_message: string;
    current_time: string;
    oidc_provider_name: string;

    constructor() {
        // Set default value
        this.with_trivy = false;
        this.admiral_endpoint = '';
        this.auth_mode = 'db_auth';
        this.primary_auth_mode = false;
        this.registry_url = '';
        this.project_creation_restriction = 'everyone';
        this.self_registration = true;
        this.has_ca_root = false;
        this.harbor_version = 'unknown';
        this.clair_vulnerability_status = {
            overall_last_update: 0,
            details: [],
        };
        this.next_scan_all = 0;
        this.registry_storage_provider_name = '';
        this.read_only = false;
        this.show_popular_repo = false;
        this.banner_message = '';
        this.oidc_provider_name = '';
    }
}
