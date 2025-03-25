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
import { ArtifactDefaultService, ArtifactService } from './artifact.service';
import { IconService } from '../../../../../../ng-swagger-gen/services/icon.service';
import { DomSanitizer } from '@angular/platform-browser';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('ArtifactService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule, HttpClientTestingModule],
            providers: [
                {
                    provide: ArtifactService,
                    useClass: ArtifactDefaultService,
                },
                IconService,
                DomSanitizer,
            ],
        });
    });

    it('should be initialized', inject(
        [ArtifactService],
        (service: ArtifactService) => {
            expect(service).toBeTruthy();
        }
    ));
});
