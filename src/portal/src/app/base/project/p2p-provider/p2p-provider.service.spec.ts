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
import { TestBed, inject } from '@angular/core/testing';
import { EXECUTION_STATUS, P2pProviderService } from './p2p-provider.service';

describe('P2pProviderService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            providers: [P2pProviderService],
        });
    });

    it('should be created', inject(
        [P2pProviderService],
        (service: P2pProviderService) => {
            expect(service).toBeTruthy();
        }
    ));
    it('function getDuration should work', inject(
        [P2pProviderService],
        (service: P2pProviderService) => {
            const date = new Date();
            const noDuration: string = service.getDuration(
                new Date(date).toUTCString(),
                new Date(date.getTime()).toUTCString()
            );
            expect(noDuration).toEqual('-');
            const durationMs: string = service.getDuration(
                new Date(date).toUTCString(),
                new Date(date.getTime() + 10).toUTCString()
            );
            expect(durationMs).toEqual('-');
            const durationSec: string = service.getDuration(
                new Date(date).toUTCString(),
                new Date(date.getTime() + 1010).toUTCString()
            );
            expect(durationSec).toEqual('1s');
            const durationMin: string = service.getDuration(
                new Date(date).toUTCString(),
                new Date(date.getTime() + 61010).toUTCString()
            );
            expect(durationMin).toEqual('1m1s');
            const durationMinOnly: string = service.getDuration(
                new Date(date).toUTCString(),
                new Date(date.getTime() + 60000).toUTCString()
            );
            expect(durationMinOnly).toEqual('1m');
        }
    ));
    it('function willChangStatus should work', inject(
        [P2pProviderService],
        (service: P2pProviderService) => {
            expect(
                service.willChangStatus(EXECUTION_STATUS.PENDING)
            ).toBeTruthy();
            expect(
                service.willChangStatus(EXECUTION_STATUS.RUNNING)
            ).toBeTruthy();
            expect(
                service.willChangStatus(EXECUTION_STATUS.SCHEDULED)
            ).toBeTruthy();
        }
    ));
});
