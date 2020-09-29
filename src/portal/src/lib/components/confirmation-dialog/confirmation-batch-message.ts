
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

export function  BathInfoChanges(batchInfo: BatchInfo, status: string, loading = false, errStatus = false, errorInfo = '') {
        batchInfo.status = status;
        batchInfo.loading = loading;
        batchInfo.errorState = errStatus;
        batchInfo.errorInfo = errorInfo;
        return batchInfo;
}

export enum BatchOperations {
    Idle,
    Delete,
    ChangeRole
}

