
/**
 * Created by pengf on 11/22/2017.
 */

export class BatchInfo {
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

export function  BathInfoChanges(list: BatchInfo, status: string, loading = false, errStatus = false, errorInfo = '') {
        list.status = status;
        list.loading = loading;
        list.errorState = errStatus;
        list.errorInfo = errorInfo;
        return list;
}

