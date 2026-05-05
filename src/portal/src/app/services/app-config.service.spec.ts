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
import { TestBed, inject, getTestBed } from '@angular/core/testing';
import {
    HttpClientTestingModule,
    HttpTestingController,
} from '@angular/common/http/testing';
import { CookieService } from 'ngx-cookie';
import { AppConfigService } from './app-config.service';
import { AppConfig } from './app-config';
import { CURRENT_BASE_HREF } from '../shared/units/utils';

describe('AppConfigService', () => {
    let injector: TestBed;
    let service: AppConfigService;
    let httpMock: HttpTestingController;
    let fakeCookieService = {
        get: function (key) {
            return null;
        },
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [
                AppConfigService,
                { provide: CookieService, useValue: fakeCookieService },
            ],
        });
        injector = getTestBed();
        service = injector.get(AppConfigService);
        httpMock = injector.get(HttpTestingController);
    });
    let systeminfo = new AppConfig();
    it('should be created', inject(
        [AppConfigService],
        (service1: AppConfigService) => {
            expect(service1).toBeTruthy();
        }
    ));

    it('load() should return data', () => {
        service.load().subscribe(res => {
            expect(res).toEqual(systeminfo);
        });

        const req = httpMock.expectOne(CURRENT_BASE_HREF + '/systeminfo');
        expect(req.request.method).toBe('GET');
        req.flush(systeminfo);
        expect(service.getConfig()).toEqual(systeminfo);
        expect(service.isLdapMode()).toBeFalsy();
        expect(service.isHttpAuthMode()).toBeFalsy();
        expect(service.isOidcMode()).toBeFalsy();
    });
});
