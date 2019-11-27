import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
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
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DevCenterComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

});
