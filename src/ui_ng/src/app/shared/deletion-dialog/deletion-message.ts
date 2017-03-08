import { DeletionTargets } from '../../shared/shared.const';

export class DeletionMessage {
    public constructor(title: string, message: string, param: string, data: any, targetId: DeletionTargets) {
        this.title = title;
        this.message = message;
        this.data = data;
        this.targetId = targetId;
        this.param = param;
    }
    title: string;
    message: string;
    data: any;
    targetId: DeletionTargets = DeletionTargets.EMPTY;
    param: string;
}