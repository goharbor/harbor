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
import { SessionService } from '../../shared/services/session.service';
import { ProjectRoutingResolver } from './project-routing-resolver.service';
import { RouterTestingModule } from '@angular/router/testing';
import { ProjectService } from '../../shared/services';

describe('ProjectRoutingResolverService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [RouterTestingModule],
            providers: [
                ProjectRoutingResolver,
                { provide: SessionService, useValue: null },
                { provide: ProjectService, useValue: null },
            ],
        });
    });

    it('should be created', inject(
        [ProjectRoutingResolver],
        (service: ProjectRoutingResolver) => {
            expect(service).toBeTruthy();
        }
    ));
});
