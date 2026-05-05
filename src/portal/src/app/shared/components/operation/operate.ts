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
export class OperateInfo {
    name: string;
    state: string;
    data: { [key: string]: string | number | boolean };
    timeStamp: number;
    timeDiff: string;
    constructor() {
        this.name = '';
        this.state = '';
        this.data = { id: -1, name: '', errorInf: '' };
        this.timeStamp = new Date().getTime();
        this.timeDiff = 'OPERATION.SECOND_AGO';
    }
}

export interface OperateInfosLocalstorage {
    updated: number; // millisecond
    data: OperateInfo[];
    newMessageCount: number;
}

export function operateChanges(
    list: OperateInfo,
    state?: string,
    errorInfo?: string,
    timeStamp?: 0
) {
    list.state = state;
    list.data.errorInf = errorInfo;
    list.timeStamp = new Date().getTime();
    return list;
}

export const OperationState = {
    progressing: 'progressing',
    success: 'success',
    failure: 'failure',
    interrupt: 'interrupt',
};
