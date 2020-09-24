import { TestBed, inject } from "@angular/core/testing";
import { TranslateService } from "@ngx-translate/core";
import { IServiceConfig, SERVICE_CONFIG } from "../entities/service.config";
import { SharedModule } from "../utils/shared/shared.module";
import { DEFAULT_LANG } from "../utils/utils";

const EN_US_LANG: any = {
  SIGN_UP: {
    TITLE: "Sign Up"
  }
};

const ZH_CN_LANG: any = {
  SIGN_UP: {
    TITLE: "注册"
  }
};

describe("TranslateService", () => {
  let testConfig: IServiceConfig = {
    langMessageLoader: "local",
    localI18nMessageVariableMap: {
      "en-us": EN_US_LANG,
      "zh-cn": ZH_CN_LANG
    }
  };
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [SharedModule],
      providers: [
        {
          provide: SERVICE_CONFIG,
          useValue: testConfig
        }
      ]
    });
  });

  it(
    "should be initialized",
    inject([TranslateService], (service: TranslateService) => {
      expect(service).toBeTruthy();
    })
  );

  it(
    "should use the specified lang",
    inject([TranslateService], (service: TranslateService) => {
      service.use(DEFAULT_LANG).subscribe(() => {
        expect(service.currentLang).toEqual(DEFAULT_LANG);
      });
    })
  );

  it(
    "should translate key to text [en-us]",
    inject([TranslateService], (service: TranslateService) => {
      service.use(DEFAULT_LANG);
      service.get("SIGN_UP.TITLE").subscribe(text => {
        expect(text).toEqual("Sign Up");
      });
    })
  );

  it(
    "should translate key to text [zh-cn]",
    inject([TranslateService], (service: TranslateService) => {
      service.use("zh-cn");
      service.get("SIGN_UP.TITLE").subscribe(text => {
        expect(text).toEqual("注册");
      });
    })
  );
});
