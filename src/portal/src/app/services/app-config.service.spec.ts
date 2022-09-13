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
        expect(service.isIntegrationMode()).toBeFalsy();
        expect(service.isLdapMode()).toBeFalsy();
        expect(service.isHttpAuthMode()).toBeFalsy();
        expect(service.isOidcMode()).toBeFalsy();
    });
});
