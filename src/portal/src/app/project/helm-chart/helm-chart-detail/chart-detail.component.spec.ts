import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { ActivatedRoute, Router } from "@angular/router";
import { SessionService } from './../../../shared/session.service';
import { of } from 'rxjs';
import { HelmChartDetailComponent } from './chart-detail.component';

describe('ChartDetailComponent', () => {
  let component: HelmChartDetailComponent;
  let fixture: ComponentFixture<HelmChartDetailComponent>;
  let fakeRouter = null;
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
      declarations: [HelmChartDetailComponent],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      imports: [
        ClarityModule,
        TranslateModule.forRoot()
      ],
      providers: [
        {
          provide: ActivatedRoute, useValue: {
            paramMap: of({ get: (key) => 'value' }),
            snapshot: {
              params: { id: 1, chart: 'chart', version: 1.0 },
              data: {
                projectResolver: {
                  role_name: 'admin'
                }
              }
            }
          }
        },
        { provide: Router, useValue: fakeRouter },
        { provide: SessionService, useValue: fakeSessionService },
        TranslateService
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(HelmChartDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
