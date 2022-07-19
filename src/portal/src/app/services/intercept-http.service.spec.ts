import { TestBed, inject } from '@angular/core/testing';
import { InterceptHttpService } from './intercept-http.service';
import { HttpRequest, HttpResponse } from '@angular/common/http';
import { of, throwError } from 'rxjs';

describe('InterceptHttpService', () => {
    const mockedCSRFToken: string = 'test';
    const mockRequest = new HttpRequest('PUT', '', {
        headers: new Map(),
    });
    const mockHandle = {
        handle: request => {
            if (request.headers.has('X-Harbor-CSRF-Token')) {
                return of(new HttpResponse({ status: 200 }));
            } else {
                return throwError(
                    new HttpResponse({
                        status: 403,
                    })
                );
            }
        },
    };
    beforeEach(() => {
        let store = {};
        spyOn(localStorage, 'getItem').and.callFake(key => {
            return store[key];
        });
        spyOn(localStorage, 'setItem').and.callFake((key, value) => {
            return (store[key] = value + '');
        });
        spyOn(localStorage, 'clear').and.callFake(() => {
            store = {};
        });
        TestBed.configureTestingModule({
            imports: [],
            providers: [InterceptHttpService],
        });
    });
    it('should be initialized', inject(
        [InterceptHttpService],
        (service: InterceptHttpService) => {
            expect(service).toBeTruthy();
        }
    ));

    it('should be get right token and send right request when the cookie not exists', inject(
        [InterceptHttpService],
        (service: InterceptHttpService) => {
            localStorage.setItem('__csrf', mockedCSRFToken);
            service.intercept(mockRequest, mockHandle).subscribe(res => {
                if (res.status === 403) {
                    expect(
                        mockRequest.headers.get('X-Harbor-CSRF-Token')
                    ).toEqual(mockedCSRFToken);
                } else {
                    expect(res.status).toEqual(200);
                }
            });
        }
    ));
});
