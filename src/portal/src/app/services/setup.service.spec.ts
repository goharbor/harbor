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
import {
    HttpClientTestingModule,
    HttpTestingController,
} from '@angular/common/http/testing';
import { SetupService } from './setup.service';

describe('SetupService', () => {
    let service: SetupService;
    let httpMock: HttpTestingController;

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [SetupService],
        });
        service = TestBed.inject(SetupService);
        httpMock = TestBed.inject(HttpTestingController);
    });

    afterEach(() => {
        httpMock.verify();
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should return setup_required=true when setup is needed', () => {
        service.clearCache();
        service.isSetupRequired().subscribe(result => {
            expect(result).toBeTrue();
        });

        const req = httpMock.expectOne('/c/setup/status');
        expect(req.request.method).toBe('GET');
        req.flush({ setup_required: true });
    });

    it('should return setup_required=false when setup is not needed', () => {
        service.clearCache();
        service.isSetupRequired().subscribe(result => {
            expect(result).toBeFalse();
        });

        const req = httpMock.expectOne('/c/setup/status');
        expect(req.request.method).toBe('GET');
        req.flush({ setup_required: false });
    });

    it('should cache the result after first call', () => {
        service.clearCache();
        service.isSetupRequired().subscribe(result => {
            expect(result).toBeTrue();
        });

        const req = httpMock.expectOne('/c/setup/status');
        req.flush({ setup_required: true });

        // Second call should use cache
        service.isSetupRequired().subscribe(result => {
            expect(result).toBeTrue();
        });

        httpMock.expectNone('/c/setup/status');
    });

    it('should POST password to /c/setup', () => {
        service.setupAdminPassword('Harbor12345').subscribe(result => {
            expect(result).toBeTruthy();
        });

        const req = httpMock.expectOne('/c/setup');
        expect(req.request.method).toBe('POST');
        expect(req.request.body).toEqual({ password: 'Harbor12345' });
        req.flush({ ok: true });
    });

    it('should clear cache on successful setup', () => {
        service.clearCache();
        // First get the status
        service.isSetupRequired().subscribe();
        const statusReq = httpMock.expectOne('/c/setup/status');
        statusReq.flush({ setup_required: true });

        // Now do setup
        service.setupAdminPassword('Harbor12345').subscribe();
        const setupReq = httpMock.expectOne('/c/setup');
        setupReq.flush({ ok: true });

        // After setup, cached status should be false
        service.isSetupRequired().subscribe(result => {
            expect(result).toBeFalse();
        });

        // No HTTP request should be made since it was cached as false
        httpMock.expectNone('/c/setup/status');
    });
});
