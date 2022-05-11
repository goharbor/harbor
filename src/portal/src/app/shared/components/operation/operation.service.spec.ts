import { OperationService } from './operation.service';
import { Subscription } from 'rxjs';
import { OperateInfo } from './operate';

let sub: Subscription;

describe('OperationService', () => {
    afterEach(() => {
        if (sub) {
            sub.unsubscribe();
            sub = null;
        }
    });
    it('should be created', () => {
        let operateInfo: OperateInfo;
        const operationService = new OperationService();
        sub = operationService.operationInfo$.subscribe(res => {
            operateInfo = res;
        });
        operationService.publishInfo(new OperateInfo());
        expect(operateInfo.timeDiff).toEqual('OPERATION.SECOND_AGO');
    });
});
