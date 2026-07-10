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
import { of } from 'rxjs';
import {
    SkipSessionRenewalService,
    skipSessionRenewal,
} from './skip-session-renewal.service';

describe('SkipSessionRenewalService', () => {
    let service: SkipSessionRenewalService;

    beforeEach(() => {
        TestBed.configureTestingModule({});
        service = TestBed.inject(SkipSessionRenewalService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should track begin and end correctly', () => {
        expect(service.shouldSkip).toBeFalse();
        service.begin();
        expect(service.shouldSkip).toBeTrue();
        service.end();
        expect(service.shouldSkip).toBeFalse();
    });

    it('should set shouldSkip via skipSessionRenewal operator', () => {
        of('test').pipe(skipSessionRenewal(service)).subscribe();
        expect(service.shouldSkip).toBeTrue();
    });
});
