import { async, ComponentFixture, TestBed, getTestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { DevCenterComponent } from './dev-center.component';
import { CookieService } from 'ngx-cookie';

describe('DevCenterComponent', () => {
  let component: DevCenterComponent;
  let fixture: ComponentFixture<DevCenterComponent>;
  const mockCookieService = {
    get: () => {
      return "xsrf";
    }
  };
  let cookie = "fdsa|ds";
  let injector: TestBed;
  let httpMock: HttpTestingController;
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [DevCenterComponent],
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
    fixture = TestBed.createComponent(DevCenterComponent);
    component = fixture.componentInstance;
    fixture.autoDetectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
  it('get swagger should return data', () => {
    const req = httpMock.expectOne('/swagger.json');
    expect(req.request.method).toBe('GET');
    req.flush({
      "host": '122.33',
    });
  });

});
