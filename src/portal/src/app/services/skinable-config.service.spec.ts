import { TestBed, inject, getTestBed } from '@angular/core/testing';
import {
    HttpClientTestingModule,
    HttpTestingController,
} from '@angular/common/http/testing';
import { SkinableConfig } from './skinable-config.service';

describe('SkinableConfig', () => {
    let injector: TestBed;
    let service: SkinableConfig;
    let httpMock: HttpTestingController;
    let product = {
        name: '',
        logo: '',
        introduction: '',
    };
    let mockCustomSkinData = {
        headerBgColor: {
            darkMode: '',
            lightMode: '',
        },
        loginBgImg: '',
        loginTitle: '',
        product: product,
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [SkinableConfig],
        });
        injector = getTestBed();
        service = injector.get(SkinableConfig);
        httpMock = injector.get(HttpTestingController);
    });

    it('should be created', inject(
        [SkinableConfig],
        (service1: SkinableConfig) => {
            expect(service1).toBeTruthy();
        }
    ));
    it('getCustomFile() should return data', () => {
        service.getCustomFile().subscribe(res => {
            expect(res).toEqual(mockCustomSkinData);
        });

        const req = httpMock.expectOne('setting.json?buildTimeStamp=0');
        expect(req.request.method).toBe('GET');
        req.flush(mockCustomSkinData);
        expect(service.getSkinConfig()).toEqual(mockCustomSkinData);
        expect(service.getSkinConfig().product).toEqual(product);
    });
});
