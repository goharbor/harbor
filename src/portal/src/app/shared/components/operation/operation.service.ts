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
