import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { SessionService } from "../shared/session.service";
import { GcPageComponent } from './gc-page.component';

describe('GcPageComponent', () => {
  let component: GcPageComponent;
  let fixture: ComponentFixture<GcPageComponent>;
  let fakeSessionService = {
    getCurrentUser: function () {
      return {
        sysadmin_flag: true,
        admin_role_in_auth: true,
      };
    }
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [GcPageComponent],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      imports: [
        ClarityModule,
        TranslateModule.forRoot()
      ],
      providers: [
        TranslateService,
        { provide: SessionService, useValue: fakeSessionService }
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(GcPageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
