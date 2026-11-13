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
import { ModelSyncExecution } from '../../../../../ng-swagger-gen/models/model-sync-execution';
import { ModelSyncTask } from '../../../../../ng-swagger-gen/models/model-sync-task';

export const MODEL_SYNC_IN_PROGRESS = 'InProgress';
export const MODEL_SYNC_PENDING = 'Pending';

export const TRIGGER_MANUAL = 'manual';
export const TRIGGER_SCHEDULED = 'scheduled';

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
