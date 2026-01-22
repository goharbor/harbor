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
import {
    ReplicationService,
    ReplicationDefaultService,
} from './replication.service';
import { SharedTestingModule } from '../shared.module';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('ReplicationService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule, HttpClientTestingModule],
            providers: [
                ReplicationDefaultService,
                {
                    provide: ReplicationService,
                    useClass: ReplicationDefaultService,
                },
            ],
        });
    });

    it('should be initialized', inject(
        [ReplicationDefaultService],
        (service: ReplicationService) => {
            expect(service).toBeTruthy();
        }
    ));
});
