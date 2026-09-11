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
import { InterceptHttpService } from './intercept-http.service';
import {
    HEADER_NO_SESSION_RENEWAL,
    SkipSessionRenewalService,
} from './skip-session-renewal.service';
import { HttpRequest, HttpResponse } from '@angular/common/http';
import { of, throwError } from 'rxjs';

describe('InterceptHttpService', () => {
    const mockRequest = new HttpRequest('PUT', '', null);
    // captures the request the interceptor forwards, so the assertions can
    // inspect the headers it added
    let forwardedRequest: HttpRequest<any>;
    const mockHandle = {
        handle: request => {
            forwardedRequest = request;
            return of(new HttpResponse({ status: 200 }));
        },
    };

    beforeEach(() => {
        forwardedRequest = null;
        TestBed.configureTestingModule({
            imports: [],
            providers: [InterceptHttpService, SkipSessionRenewalService],
        });
    });

    it('should be initialized', inject(
        [InterceptHttpService],
        (service: InterceptHttpService) => {
            expect(service).toBeTruthy();
        }
    ));

    it('should not set the no-session-renewal header by default', inject(
        [InterceptHttpService],
        (service: InterceptHttpService) => {
            service.intercept(mockRequest, mockHandle).subscribe();
            expect(
                forwardedRequest.headers.has(HEADER_NO_SESSION_RENEWAL)
            ).toBe(false);
        }
    ));

    it('should set the no-session-renewal header when the request is flagged', inject(
        [InterceptHttpService, SkipSessionRenewalService],
        (
            service: InterceptHttpService,
            skipSessionRenewalService: SkipSessionRenewalService
        ) => {
            skipSessionRenewalService.begin();
            service.intercept(mockRequest, mockHandle).subscribe();
            expect(
                forwardedRequest.headers.get(HEADER_NO_SESSION_RENEWAL)
            ).toEqual('true');
            // the flag is consumed by the interceptor, so it must not leak to the next request
            expect(skipSessionRenewalService.shouldSkip).toBe(false);
        }
    ));

    it('should convert a 504 error into json format', inject(
        [InterceptHttpService],
        (service: InterceptHttpService) => {
            const failingHandle = {
                handle: () => throwError(() => ({ status: 504 })),
            };
            let caught: any = null;
            service.intercept(mockRequest, failingHandle).subscribe({
                error: error => (caught = error),
            });
            expect(caught).toBeTruthy();
            expect(caught.status).toEqual(504);
            expect(caught.error).toEqual('504 gateway timeout');
        }
    ));

    it('should rethrow non-504 errors unchanged', inject(
        [InterceptHttpService],
        (service: InterceptHttpService) => {
            const failingHandle = {
                handle: () =>
                    throwError(() => ({ status: 500, error: 'server error' })),
            };
            let caught: any = null;
            service.intercept(mockRequest, failingHandle).subscribe({
                error: error => (caught = error),
            });
            expect(caught).toBeTruthy();
            expect(caught.status).toEqual(500);
            expect(caught.error).toEqual('server error');
        }
    ));
});
