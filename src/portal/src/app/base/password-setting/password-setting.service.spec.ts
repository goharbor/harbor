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
import { PasswordSettingService } from './password-setting.service';
import { CURRENT_BASE_HREF } from '../../shared/units/utils';

describe('PasswordSettingService', () => {
    let injector: TestBed;
    let service: PasswordSettingService;
    let httpMock: HttpTestingController;
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [PasswordSettingService],
        });
        injector = getTestBed();
        service = injector.get(PasswordSettingService);
        httpMock = injector.get(HttpTestingController);
    });

    it('should be created', inject(
        [PasswordSettingService],
        (service1: PasswordSettingService) => {
            expect(service1).toBeTruthy();
        }
    ));

    const mockPasswordSetting = {
        old_password: 'string',
        new_password: 'string1',
    };

    it('changePassword() should success', () => {
        service.changePassword(1, mockPasswordSetting).subscribe(res => {
            expect(res).toEqual(null);
        });

        const req = httpMock.expectOne(CURRENT_BASE_HREF + '/users/1/password');
        expect(req.request.method).toBe('PUT');
        req.flush(null);
    });
    it('sendResetPasswordMail() should return data', () => {
        service.sendResetPasswordMail('123').subscribe(res => {
            expect(res).toEqual(null);
        });

        const req = httpMock.expectOne('/c/sendEmail?email=123');
        expect(req.request.method).toBe('GET');
        req.flush(null);
    });
    it('resetPassword() should return data', () => {
        service.resetPassword('1234', 'Harbor12345').subscribe(res => {
            expect(res).toEqual(null);
        });

        const req = httpMock.expectOne('/c/reset');
        expect(req.request.method).toBe('POST');
        req.flush(null);
    });
});
