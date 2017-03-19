import { ConfirmationTargets } from '../../shared/shared.const';

export class ConfirmationMessage {
    public constructor(title: string, message: string, param: string, data: any, targetId: ConfirmationTargets) {
        this.title = title;
        this.message = message;
        this.data = data;
        this.targetId = targetId;
        this.param = param;
    }
    title: string;
    message: string;
    data: any = {};//default is empty
    targetId: ConfirmationTargets = ConfirmationTargets.EMPTY;
    param: string;
}