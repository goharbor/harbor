import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ActivatedRoute } from '@angular/router';
import { of } from 'rxjs';
import { AppConfigService } from "../../services/app-config.service";
import { SummaryComponent } from './summary.component';
import { ProjectService, UserPermissionService } from "../../../lib/services";
import { ErrorHandler } from "../../../lib/utils/error-handler";


describe('SummaryComponent', () => {
  let component: SummaryComponent;
  let fixture: ComponentFixture<SummaryComponent>;
  let fakeAppConfigService = null;
  let fakeProjectService = {
    getProjectSummary: function () {
      return of();
    }
  };
  let fakeErrorHandler = null;
  let fakeUserPermissionService = {
    hasProjectPermissions: function() {
      return of([true, true]);
    }
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [SummaryComponent],
      imports: [
        ClarityModule,
        TranslateModule.forRoot()
      ],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      providers: [
        TranslateService,
        { provide: AppConfigService, useValue: fakeAppConfigService },
        { provide: ProjectService, useValue: fakeProjectService },
        { provide: ErrorHandler, useValue: fakeErrorHandler },
        { provide: UserPermissionService, useValue: fakeUserPermissionService },
        {
          provide: ActivatedRoute, useValue: {
            paramMap: of({ get: (key) => 'value' }),
            snapshot: {
              parent: {
                params: { id: 1 }
              },
              data: 1
            }
          }
        },
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(SummaryComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
