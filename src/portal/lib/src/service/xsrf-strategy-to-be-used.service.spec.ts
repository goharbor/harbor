import { TestBed, inject } from '@angular/core/testing';

import { SharedModule } from '../shared/shared.module';
import { CookieService } from "ngx-cookie";
import { XSRFStrategyToBeUsed } from './xsrf-strategy-to-be-used.service';

describe('XSRFStrategyToBeUsed', () => {
    let cookie = "fdsa|ds";
    let req = { headers: new Map()};
    let mockCookieService = {
        get: function () {
            return cookie;
        },
        set: function (cookieStr: string) {
            cookie = cookieStr;
        }
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [
                SharedModule
            ],
            providers: [
                {provide: XSRFStrategyToBeUsed, useClass: XSRFStrategyToBeUsed, deps: [ CookieService]},
                { provide: CookieService, useValue: mockCookieService }
            ]
        });

    });

    it('should be initialized', inject([XSRFStrategyToBeUsed], (service: XSRFStrategyToBeUsed) => {
        expect(service).toBeTruthy();
    }));

    it('should be get right token when the cookie exists', inject([XSRFStrategyToBeUsed],
        (service: XSRFStrategyToBeUsed) => {
            mockCookieService.set("fdsa|ds");
            service.configureRequest(req);
            expect(btoa(req.headers.get(service.headerName))).toEqual(cookie.split("|")[0]);
        }));

    it('should be get right token when the cookie does not exist', inject([XSRFStrategyToBeUsed],
        (service: XSRFStrategyToBeUsed) => {
            mockCookieService.set(null);
            service.configureRequest(req);
            expect(req.headers.get(service.headerName)).toBeNull();
        }));

});
