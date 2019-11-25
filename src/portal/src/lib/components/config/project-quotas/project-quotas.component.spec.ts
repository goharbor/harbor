import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { ProjectQuotasComponent } from './project-quotas.component';
import { IServiceConfig, SERVICE_CONFIG } from '../../../entities/service.config';
import { SharedModule } from '../../../utils/shared/shared.module';
import { RouterModule } from '@angular/router';
import { EditProjectQuotasComponent } from './edit-project-quotas/edit-project-quotas.component';
import { InlineAlertComponent } from '../../inline-alert/inline-alert.component';
import {
  ConfigurationService, ConfigurationDefaultService, QuotaService
  , QuotaDefaultService, Quota, RequestQueryParams
} from '../../../services';
import { ErrorHandler } from '../../../utils/error-handler';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import {APP_BASE_HREF} from '@angular/common';
describe('ProjectQuotasComponent', () => {
  let spy: jasmine.Spy;
  let quotaService: QuotaService;

  let component: ProjectQuotasComponent;
  let fixture: ComponentFixture<ProjectQuotasComponent>;

  let config: IServiceConfig = {
    quotaUrl: "/api/quotas/testing"
  };
  let mockQuotaList: Quota[] = [{
    id: 1111,
    ref: {
      id: 1111,
      name: "project1",
      owner_name: "project1"
    },
    creation_time: "12212112121",
    update_time: "12212112121",
      hard: {
        count: -1,
        storage: -1,
      },
      used: {
        count: 1234,
        storage: 1234
      },
  }
  ];
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        RouterModule.forRoot([])
      ],
      declarations: [ProjectQuotasComponent, EditProjectQuotasComponent, InlineAlertComponent],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: ConfigurationService, useClass: ConfigurationDefaultService },
        { provide: QuotaService, useClass: QuotaDefaultService },
        { provide: APP_BASE_HREF, useValue : '/' }

      ]
    })
      .compileComponents();
  }));

  beforeEach(async(() => {

    fixture = TestBed.createComponent(ProjectQuotasComponent);
    component = fixture.componentInstance;
    component.quotaHardLimitValue = {
      countLimit: 1111,
      storageLimit: 23,
      storageUnit: 'GB'
    };
    component.loading = true;
    quotaService = fixture.debugElement.injector.get(QuotaService);
    spy = spyOn(quotaService, 'getQuotaList')
      .and.callFake(function (params: RequestQueryParams) {
        let header = new Map();
        header.set("X-Total-Count", 123);
        const httpRes = {
          headers: header,
          body: mockQuotaList
        };
        return of(httpRes).pipe(delay(0));
      });

    fixture.detectChanges();
  }));

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
