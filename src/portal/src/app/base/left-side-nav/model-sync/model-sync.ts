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
import { Registry } from '../../../../../ng-swagger-gen/models/registry';
import { ReplicationExecution } from '../../../../../ng-swagger-gen/models/replication-execution';
import { ReplicationFilter } from '../../../../../ng-swagger-gen/models/replication-filter';
import { ReplicationPolicy } from '../../../../../ng-swagger-gen/models/replication-policy';
import { ReplicationTask } from '../../../../../ng-swagger-gen/models/replication-task';

export const MODEL_SYNC_IN_PROGRESS = 'InProgress';
export const MODEL_SYNC_PENDING = 'Pending';

export const TRIGGER_MANUAL = 'manual';
export const TRIGGER_SCHEDULED = 'scheduled';

// A replication tag filter without a value copies every branch and tag of the model.
export const DEFAULT_REVISION = 'main';

const FILTER_NAME = 'name';
const FILTER_TAG = 'tag';
const FILTER_FILE = 'file';

// ModelSyncPolicy is the Model Sync view of a replication rule whose source registry
// supports previews: one model, one revision, a file filter, one destination project.
export interface ModelSyncPolicy {
    id?: number;
    name: string;
    description?: string;
    enabled: boolean;
    registry_id: number;
    registry?: Registry;
    src_repository: string;
    src_revision?: string;
    file_filters?: string[];
    file_excludes?: string[];
    dest_project_name: string;
    dest_repository?: string;
    trigger: { type: string; trigger_settings?: { cron?: string } };
    creation_time?: string;
    update_time?: string;
}

export type ModelSyncExecution = ReplicationExecution;
export type ModelSyncTask = ReplicationTask;

export function isExecutionInProgress(execution: ModelSyncExecution): boolean {
    return !!execution && execution.status === MODEL_SYNC_IN_PROGRESS;
}

export function isTaskInProgress(task: ModelSyncTask): boolean {
    return (
        !!task &&
        (task.status === MODEL_SYNC_IN_PROGRESS ||
            task.status === MODEL_SYNC_PENDING)
    );
}

// convert "a, b\n c" style user input into a filter array
export function parseFileFilters(input: string): string[] {
    if (!input) {
        return [];
    }
    return input
        .split(/[\n,]/)
        .map(item => item.trim())
        .filter(item => item.length > 0);
}

export function formatFileFilters(filters: string[]): string {
    return (filters || []).join(', ');
}

export function toReplicationPolicy(p: ModelSyncPolicy): ReplicationPolicy {
    const filters: ReplicationFilter[] = [
        { type: FILTER_NAME, value: p.src_repository },
        {
            type: FILTER_TAG,
            value: p.src_revision || DEFAULT_REVISION,
            decoration: 'matches',
        },
    ];
    if (p.file_filters?.length || p.file_excludes?.length) {
        filters.push({
            type: FILTER_FILE,
            value: {
                include: p.file_filters || [],
                exclude: p.file_excludes || [],
            },
        });
    }
    return {
        name: p.name,
        description: p.description || '',
        src_registry: { id: p.registry_id },
        dest_namespace: p.dest_project_name,
        // -1 puts the model below the project whatever its depth, e.g. "gpt2" and "org/model"
        dest_namespace_replace_count: -1,
        dest_repository: p.dest_repository || '',
        filters,
        trigger: p.trigger,
        enabled: p.enabled,
        override: true,
        deletion: false,
        revision_tag: true,
        single_active_replication: true,
    };
}

export function fromReplicationPolicy(r: ReplicationPolicy): ModelSyncPolicy {
    const filter = (type: string) =>
        (r.filters || []).find(f => f.type === type)?.value;
    const file = filter(FILTER_FILE) as {
        include?: string[];
        exclude?: string[];
    };
    return {
        id: r.id,
        name: r.name,
        description: r.description,
        enabled: r.enabled,
        registry_id: r.src_registry?.id,
        registry: r.src_registry,
        src_repository: (filter(FILTER_NAME) as string) || '',
        src_revision: (filter(FILTER_TAG) as string) || '',
        file_filters: file?.include || [],
        file_excludes: file?.exclude || [],
        dest_project_name: r.dest_namespace,
        dest_repository: r.dest_repository || '',
        trigger: {
            type: r.trigger?.type || TRIGGER_MANUAL,
            trigger_settings: r.trigger?.trigger_settings,
        },
        creation_time: r.creation_time,
        update_time: r.update_time,
    };
}

// destRepository is where a copy lands: the project, then the chosen name or the last
// path component of the model ID, matching the replication namespace rewrite.
export function destRepository(p: {
    dest_project_name?: string;
    dest_repository?: string;
    src_repository?: string;
}): string {
    const name =
        (p.dest_repository || '').trim() ||
        (p.src_repository || '').trim().split('/').pop().toLowerCase();
    return p.dest_project_name && name ? `${p.dest_project_name}/${name}` : '';
}
