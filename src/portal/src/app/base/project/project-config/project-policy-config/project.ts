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
export class Project {
    project_id: number;
    owner_id?: number;
    registry_id?: number | null;
    name: string;
    creation_time?: Date | string;
    deleted?: number;
    owner_name?: string;
    togglable?: boolean;
    update_time?: Date | string;
    current_user_role_id?: number;
    repo_count?: number;
    has_project_admin_role?: boolean;
    is_member?: boolean;
    role_name?: string;
    metadata?: {
        public: string | boolean;
        enable_content_trust?: string | boolean;
        enable_content_trust_cosign?: string | boolean;
        prevent_vul: string | boolean;
        severity: string;
        auto_scan: string | boolean;
        auto_sbom_generation: string | boolean;
        reuse_sys_cve_allowlist?: string;
        proxy_speed_kb?: number | null;
    };
    cve_allowlist?: object;
    constructor() {
        this.metadata.public = false;
        this.metadata.enable_content_trust_cosign = false;
        this.metadata.prevent_vul = false;
        this.metadata.severity = 'low';
        this.metadata.auto_scan = false;
        this.metadata.auto_sbom_generation = false;
        this.metadata.proxy_speed_kb = -1;
    }
}
