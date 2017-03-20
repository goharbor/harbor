import { ConfirmationState, ConfirmationTargets } from '../shared.const';

export class ConfirmationAcknowledgement {
    constructor(state: ConfirmationState, data: any, source: ConfirmationTargets) {
        this.state = state;
        this.data = data;
        this.source = source;
    }

    state: ConfirmationState = ConfirmationState.NA;
    data: any = {};
    source: ConfirmationTargets = ConfirmationTargets.EMPTY;
}