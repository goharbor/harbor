export class OperateInfo {
    name: string;
    state: string;
    data: { [key: string]: string | number };
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
