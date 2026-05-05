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
import { Subject } from 'rxjs';
import { OperateInfo } from './operate';

@Injectable({
    providedIn: 'root',
})
export class OperationService {
    subjects: Subject<any> = null;

    operationInfoSource = new Subject<OperateInfo>();
    operationInfo$ = this.operationInfoSource.asObservable();

    publishInfo(data: OperateInfo): void {
        this.operationInfoSource.next(data);
    }
}

export function downloadCVEs(data, filename) {
    let url = window.URL.createObjectURL(data);
    let a = document.createElement('a');
    document.body.appendChild(a);
    a.setAttribute('style', 'display: none');
    a.href = url;
    a.download = filename;
    a.click();
    window.URL.revokeObjectURL(url);
    a.remove();
}
export enum EventState {
    SUCCESS = 'success',
    FAILURE = 'failure',
    INTERRUPT = 'interrupt',
    PROGRESSING = 'progressing',
}

export enum ExportJobStatus {
    PENDING = 'Pending',
    RUNNING = 'Running',
    STOPPED = 'Stopped',
    ERROR = 'Error',
    SUCCESS = 'Success',
    SCHEDULED = 'Scheduled',
}
