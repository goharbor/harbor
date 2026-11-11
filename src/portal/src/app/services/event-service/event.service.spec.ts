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
import { TestBed } from '@angular/core/testing';
import { EventService, HarborEvent } from './event.service';
import { Subscription } from 'rxjs';

describe('EventServiceService', () => {
    let service: EventService;

    beforeEach(() => {
        TestBed.configureTestingModule({});
        service = TestBed.inject(EventService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('able to subscribe', () => {
        let result: string;
        const sub1 = service.subscribe(HarborEvent.SCROLL, data => {
            result = data;
        });
        expect(sub1).toBeTruthy();
        expect(sub1 instanceof Subscription).toEqual(true);
        service.publish(HarborEvent.SCROLL, 'resultString');
        sub1.unsubscribe();
        expect(result).toEqual('resultString');
    });
});
