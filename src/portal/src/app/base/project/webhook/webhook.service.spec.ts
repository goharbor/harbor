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
import { inject, TestBed } from '@angular/core/testing';
import { ProjectWebhookService } from './webhook.service';

describe('WebhookService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            providers: [ProjectWebhookService],
        });
    });

    it('should be created', inject(
        [ProjectWebhookService],
        (service: ProjectWebhookService) => {
            expect(service).toBeTruthy();
        }
    ));
    it('function eventTypeToText should work', inject(
        [ProjectWebhookService],
        (service: ProjectWebhookService) => {
            expect(service).toBeTruthy();
            const eventType: string = 'REPLICATION';
            expect(service.eventTypeToText(eventType)).toEqual(
                'Replication status changed'
            );
            const mockedEventType: string = 'TEST';
            expect(service.eventTypeToText(mockedEventType)).toEqual('TEST');
        }
    ));
});
