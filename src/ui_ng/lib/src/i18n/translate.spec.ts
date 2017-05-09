import { TestBed, inject } from '@angular/core/testing';
import { SharedModule } from '../shared/shared.module';
import { TranslateService } from '@ngx-translate/core';
import { DEFAULT_LANG } from '../utils';

describe('TranslateService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      providers: []
    });
  });

  it('should be initialized', inject([TranslateService], (service: TranslateService) => {
    expect(service).toBeTruthy();
  }));

  it('should use the specified lang', inject([TranslateService], (service: TranslateService) => {
    service.use(DEFAULT_LANG).subscribe(() => {
      expect(service.currentLang).toEqual(DEFAULT_LANG);
    });
  }));

  it('should translate key to text [en-us]', inject([TranslateService], (service: TranslateService) => {
    service.use(DEFAULT_LANG);
    service.get('APP_TITLE.HARBOR').subscribe(text => {
      expect(text).toEqual('Harbor');
    });
  }));

  it('should translate key to text [zh-cn]', inject([TranslateService], (service: TranslateService) => {
    service.use('zh-cn');
    service.get('SIGN_UP.TITLE').subscribe(text => {
      expect(text).toEqual('注册');
    });
  }));

});
