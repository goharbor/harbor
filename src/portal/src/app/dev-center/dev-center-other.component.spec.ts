import { async, ComponentFixture, TestBed, getTestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { DevCenterOtherComponent } from './dev-center-other.component';
import { CookieService } from 'ngx-cookie';

describe('DevCenterOtherComponent', () => {
  let component: DevCenterOtherComponent;
  let fixture: ComponentFixture<DevCenterOtherComponent>;
  const mockCookieService = {
    get: () => {
      return "xsrf";
    }
  };
  let injector: TestBed;
  let httpMock: HttpTestingController;
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [DevCenterOtherComponent],
      imports: [
        HttpClientTestingModule,
        TranslateModule.forRoot()
      ],
      providers: [
        TranslateService,
        {
          provide: CookieService, useValue: mockCookieService
        }
      ],
    })
      .compileComponents();
    injector = getTestBed();
    httpMock = injector.get(HttpTestingController);
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DevCenterOtherComponent);
    component = fixture.componentInstance;
    fixture.autoDetectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
  it('get swagger should return data', () => {
    const req = httpMock.expectOne('/swagger3.json');
    expect(req.request.method).toBe('GET');
    req.flush({
      "host": '122.33',
    });
  });

});
