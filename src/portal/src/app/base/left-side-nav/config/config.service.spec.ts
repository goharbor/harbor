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
import { ConfigureService } from 'ng-swagger-gen/services/configure.service';
import { of } from 'rxjs';
import { SharedTestingModule } from '../../../shared/shared.module';
import { Configuration } from './config';
import { ConfigService } from './config.service';

describe('ConfigService', () => {
    const fakedConfigureService = {
        getConfigurations(): any {
            return of(null);
        },
    };
    let getConfigSpy: jasmine.Spy;
    beforeEach(() => {
        getConfigSpy = spyOn(
            fakedConfigureService,
            'getConfigurations'
        ).and.returnValue(of(new Configuration()));
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            providers: [
                ConfigService,
                { provide: ConfigureService, useValue: fakedConfigureService },
            ],
        });
    });

    it('should be created', inject(
        [ConfigService],
        (service: ConfigService) => {
            expect(service).toBeTruthy();
        }
    ));

    it('should update config', inject(
        [ConfigService],
        (service: ConfigService) => {
            expect(getConfigSpy.calls.count()).toEqual(0);
            service.updateConfig();
            expect(getConfigSpy.calls.count()).toEqual(1);
            // update again
            service.updateConfig();
            expect(getConfigSpy.calls.count()).toEqual(2);
            expect(service).toBeTruthy();
        }
    ));
});
