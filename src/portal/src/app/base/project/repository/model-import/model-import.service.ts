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

import { Injectable } from '@angular/core';
import { HttpClient, HttpResponse } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface ModelImportPolicy {
    id?: number;
    name: string;
    description?: string;
    project_id?: number;
    target_repository: string;
    target_tag: string;
    hf_repo_id: string;
    hf_revision: string;
    hf_subpath?: string;
    hf_token?: string;
    last_synced_commit?: string;
    last_artifact_digest?: string;
    creator?: string;
    enabled: boolean;
    trigger: string;
    extra_attrs?: string;
    creation_time?: string;
    update_time?: string;
}

export interface ModelImportValidationRequest {
    hf_repo_id: string;
    hf_revision?: string;
    hf_subpath?: string;
    hf_token?: string;
}

export interface ModelImportExecutionRequest extends ModelImportValidationRequest {
    target_repository: string;
    target_tag: string;
    last_synced_commit?: string;
    last_artifact_digest?: string;
}

export interface ModelImportValidation {
    hf_repo_id: string;
    hf_revision: string;
    commit_sha: string;
    source_url: string;
    file_count: number;
    card_data?: { [key: string]: unknown };
}

@Injectable()
export class ModelImportService {
    constructor(private http: HttpClient) {}

    createPolicy(
        projectName: string,
        policy: ModelImportPolicy
    ): Observable<HttpResponse<unknown>> {
        return this.http.post(
            `/api/v2.0/projects/${encodeURIComponent(
                projectName
            )}/model-import/policies`,
            policy,
            { observe: 'response' }
        );
    }

    runPolicy(
        projectName: string,
        policyName: string
    ): Observable<HttpResponse<unknown>> {
        return this.http.post(
            `/api/v2.0/projects/${encodeURIComponent(
                projectName
            )}/model-import/policies/${encodeURIComponent(
                policyName
            )}/executions`,
            {},
            { observe: 'response' }
        );
    }

    runImport(
        projectName: string,
        request: ModelImportExecutionRequest
    ): Observable<HttpResponse<unknown>> {
        return this.http.post(
            `/api/v2.0/projects/${encodeURIComponent(
                projectName
            )}/model-import/executions`,
            request,
            { observe: 'response' }
        );
    }

    validateSource(
        projectName: string,
        request: ModelImportValidationRequest
    ): Observable<ModelImportValidation> {
        return this.http.post<ModelImportValidation>(
            `/api/v2.0/projects/${encodeURIComponent(
                projectName
            )}/model-import/validate`,
            request
        );
    }
}
