import { Injectable } from '@angular/core';
import { Subject } from 'rxjs/Subject';
// tslint:disable-next-line:no-unused-variable
import { Observable } from "rxjs/Observable";
import {OperateInfo} from "./operate";

@Injectable()
export class OperationService {
    subjects: Subject<any> = null;

    operationInfoSource = new Subject<OperateInfo>();
    operationInfo$ = this.operationInfoSource.asObservable();

    publishInfo(data: OperateInfo): void {
        this.operationInfoSource.next(data);
    }
}
