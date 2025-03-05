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
import { RouterTestingModule } from '@angular/router/testing';
import { ReplicationTasksRoutingResolverService } from './replication-tasks-routing-resolver.service';
import { ReplicationService } from '../../../../ng-swagger-gen/services';

describe('ReplicationTasksRoutingResolverService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [RouterTestingModule],
            providers: [{ provide: ReplicationService, useValue: null }],
        });
    });

    it('should be created', inject(
        [ReplicationTasksRoutingResolverService],
        (service: ReplicationTasksRoutingResolverService) => {
            expect(service).toBeTruthy();
        }
    ));
});
