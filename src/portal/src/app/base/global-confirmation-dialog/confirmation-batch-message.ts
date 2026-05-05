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

/**
 * Created by pengf on 11/22/2017.
 */

export class BatchInfo {
    id?: number;
    name: string;
    status: string;
    loading: boolean;
    errorState: boolean;
    errorInfo: string;
    constructor() {
        this.status = 'pending';
        this.loading = false;
        this.errorState = false;
        this.errorInfo = '';
    }
}

export function BathInfoChanges(
    batchInfo: BatchInfo,
    status: string,
    loading = false,
    errStatus = false,
    errorInfo = ''
) {
    batchInfo.status = status;
    batchInfo.loading = loading;
    batchInfo.errorState = errStatus;
    batchInfo.errorInfo = errorInfo;
    return batchInfo;
}

export enum BatchOperations {
    Idle,
    Delete,
    ChangeRole,
}
