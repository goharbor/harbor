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
import { Observable, of } from 'rxjs';
import { map, shareReplay, switchMap } from 'rxjs/operators';
import { ReplicationService } from '../../../../../ng-swagger-gen/services/replication.service';
import { RegistryService } from '../../../../../ng-swagger-gen/services/registry.service';
import { Registry } from '../../../../../ng-swagger-gen/models/registry';
import { RegistryPreview } from '../../../../../ng-swagger-gen/models/registry-preview';
import { CURRENT_BASE_HREF } from '../../../shared/units/utils';
import {
    fromReplicationPolicy,
    ModelSyncExecution,
    ModelSyncPolicy,
    ModelSyncTask,
    toReplicationPolicy,
} from './model-sync';

// registries are listed with a large page; a Harbor rarely has more model hubs than that
const REGISTRY_PAGE_SIZE = 100;

export interface Page<T> {
    items: T[];
    total: number;
}

export interface PageQuery {
    page: number;
    pageSize: number;
    sort?: string;
}

function total(headers: { get(name: string): string | null }): number {
    const value = headers.get('X-Total-Count');
    return value ? parseInt(value, 10) : 0;
}

// ModelSyncService presents the replication rules whose source registry supports previews
// as model sync policies. It owns no API of its own: rules, executions, tasks and logs are
// the replication ones, the preview is POST /registries/{id}/preview.
@Injectable({ providedIn: 'root' })
export class ModelSyncService {
    private sourceTypes$: Observable<string[]>;

    constructor(
        private replicationService: ReplicationService,
        private registryService: RegistryService
    ) {}

    // sourceTypes lists the registry types that can be a model sync source.
    sourceTypes(): Observable<string[]> {
        if (!this.sourceTypes$) {
            this.sourceTypes$ = this.registryService
                .listRegistryProviderInfos()
                .pipe(
                    map(infos =>
                        Object.keys(infos || {})
                            .filter(type => infos[type]?.supported_preview)
                            .sort()
                    ),
                    shareReplay(1)
                );
        }
        return this.sourceTypes$;
    }

    isSourceType(type: string): Observable<boolean> {
        return this.sourceTypes().pipe(map(types => types.includes(type)));
    }

    sourceRegistries(): Observable<Registry[]> {
        return this.sourceTypes().pipe(
            switchMap(types =>
                this.registryService
                    .listRegistries({ pageSize: REGISTRY_PAGE_SIZE })
                    .pipe(
                        map(registries =>
                            (registries || []).filter(r =>
                                types.includes(r.type)
                            )
                        )
                    )
            )
        );
    }

    listPolicies(
        query: PageQuery,
        name?: string
    ): Observable<Page<ModelSyncPolicy>> {
        return this.sourceRegistries().pipe(
            switchMap(registries => {
                if (!registries.length) {
                    return of({ items: [], total: 0 });
                }
                const ids = registries.map(r => r.id).join(' ');
                let q = `src_registry_id={${ids}}`;
                if (name) {
                    q += `,name=~${name}`;
                }
                return this.replicationService
                    .listReplicationPoliciesResponse({
                        q: encodeURIComponent(q),
                        page: query.page,
                        pageSize: query.pageSize,
                        sort: query.sort,
                    })
                    .pipe(
                        map(res => ({
                            items: (res.body || []).map(fromReplicationPolicy),
                            total: total(res.headers),
                        }))
                    );
            })
        );
    }

    getPolicy(id: number): Observable<ModelSyncPolicy> {
        return this.replicationService
            .getReplicationPolicy({ id })
            .pipe(map(fromReplicationPolicy));
    }

    createPolicy(policy: ModelSyncPolicy): Observable<null> {
        return this.replicationService.createReplicationPolicy({
            policy: toReplicationPolicy(policy),
        });
    }

    updatePolicy(id: number, policy: ModelSyncPolicy): Observable<null> {
        return this.replicationService.updateReplicationPolicy({
            id,
            policy: toReplicationPolicy(policy),
        });
    }

    deletePolicy(id: number): Observable<null> {
        return this.replicationService.deleteReplicationPolicy({ id });
    }

    start(policyId: number): Observable<null> {
        return this.replicationService.startReplication({
            execution: { policy_id: policyId },
        });
    }

    stop(executionId: number): Observable<null> {
        return this.replicationService.stopReplication({ id: executionId });
    }

    listExecutions(
        policyId: number,
        query: PageQuery,
        filter: { trigger?: string; status?: string } = {}
    ): Observable<Page<ModelSyncExecution>> {
        return this.replicationService
            .listReplicationExecutionsResponse({
                policyId,
                page: query.page,
                pageSize: query.pageSize,
                sort: query.sort,
                trigger: filter.trigger,
                status: filter.status,
            })
            .pipe(
                map(res => ({
                    items: res.body || [],
                    total: total(res.headers),
                }))
            );
    }

    getExecution(executionId: number): Observable<ModelSyncExecution> {
        return this.replicationService.getReplicationExecution({
            id: executionId,
        });
    }

    listTasks(
        executionId: number,
        query: PageQuery
    ): Observable<Page<ModelSyncTask>> {
        return this.replicationService
            .listReplicationTasksResponse({
                id: executionId,
                page: query.page,
                pageSize: query.pageSize,
                sort: query.sort,
            })
            .pipe(
                map(res => ({
                    items: res.body || [],
                    total: total(res.headers),
                }))
            );
    }

    taskLogUrl(executionId: number, taskId: number): string {
        return `${CURRENT_BASE_HREF}/replication/executions/${executionId}/tasks/${taskId}/log`;
    }

    preview(
        registryId: number,
        repository: string,
        reference: string,
        include: string[],
        exclude: string[]
    ): Observable<RegistryPreview> {
        return this.registryService.previewRegistryArtifact({
            id: registryId,
            preview: {
                repository,
                reference,
                file_filter: { include, exclude },
            },
        });
    }
}
