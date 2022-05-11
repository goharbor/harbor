import { Injectable } from '@angular/core';
import { Observable, Subject } from 'rxjs';
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
