// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
