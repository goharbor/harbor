/**
 * Created by pengf on 5/11/2018.
 */

import {Type} from "@angular/core";
import {OperationComponent} from "./operation.component";

export * from  "./operation.component";
export * from './operate';
export * from './operation.service';
export const OPERATION_DIRECTIVES: Type<any>[] = [
    OperationComponent
];
