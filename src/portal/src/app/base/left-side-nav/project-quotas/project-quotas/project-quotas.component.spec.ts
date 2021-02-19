import { waitForAsync, ComponentFixture, TestBed } from '@angular/core/testing';
import { ProjectQuotasComponent } from './project-quotas.component';
import { Router } from '@angular/router';
import {
  QuotaService
  , QuotaDefaultService, Quota, RequestQueryParams
} from '../../../../shared/services';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { APP_BASE_HREF } from '@angular/common';
import { SharedTestingModule } from "../../../../shared/shared.module";
import { EditProjectQuotasComponent } from "./edit-project-quotas/edit-project-quotas.component";


describe('ProjectQuotasComponent', () => {
  let spy: jasmine.Spy;
  let spyUpdate: jasmine.Spy;
  let spyRoute: jasmine.Spy;
  let quotaService: QuotaService;
  let component: ProjectQuotasComponent;
  let fixture: ComponentFixture<ProjectQuotasComponent>;
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
        storage: -1,
      },
      used: {
        storage: 1234
      },
  }
  ];
  const fakedRouter = {
    navigate() {
      return undefined;
    }
  };
  const fakedErrorHandler = {
    error() {
      return undefined;
    },
    info() {
      return undefined;
    }
  };
  const timeout = (ms: number) => {
     return new Promise(resolve => setTimeout(resolve, ms));
  };
  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedTestingModule,
      ],
      declarations: [
        ProjectQuotasComponent,
        EditProjectQuotasComponent
      ],
      providers: [
        { provide: ErrorHandler, useValue: fakedErrorHandler },
        { provide: QuotaService, useClass: QuotaDefaultService },
        { provide: APP_BASE_HREF, useValue : '/' },
        { provide: Router, useValue: fakedRouter }
      ]
    })
      .compileComponents();
  }));

  beforeEach(waitForAsync(() => {

    fixture = TestBed.createComponent(ProjectQuotasComponent);
    component = fixture.componentInstance;
    component.quotaHardLimitValue = {
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
    spyUpdate = spyOn(quotaService, 'updateQuota').and.returnValue(of(null));
    spyRoute = spyOn(fixture.debugElement.injector.get(Router), 'navigate').and.returnValue(of(null));
    fixture.detectChanges();
  }));

  it('should create', () => {
    expect(component).toBeTruthy();
  });
  it('should open edit quota modal', async () => {
    // wait getting list and rendering
    await timeout(10);
    fixture.detectChanges();
    await fixture.whenStable();
    const openEditButton: HTMLButtonElement = fixture.nativeElement.querySelector("#open-edit");
    openEditButton.dispatchEvent(new Event("click"));
    fixture.detectChanges();
    await fixture.whenStable();
    const modal: HTMLElement = fixture.nativeElement.querySelector("clr-modal");
    expect(modal).toBeTruthy();
  });
  it('should call navigate function', async () => {
    // wait getting list and rendering
    await timeout(10);
    fixture.detectChanges();
    await fixture.whenStable();
    const a: HTMLElement = fixture.nativeElement.querySelector('clr-dg-cell a');
    a.dispatchEvent(new Event("click"));
    expect(spyRoute.calls.count()).toEqual(1);
  });
});
