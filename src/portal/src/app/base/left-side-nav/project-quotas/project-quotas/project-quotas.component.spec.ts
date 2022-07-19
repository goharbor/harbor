import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ProjectQuotasComponent } from './project-quotas.component';
import { Router } from '@angular/router';
import { Quota } from '../../../../shared/services';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { Observable, of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { APP_BASE_HREF } from '@angular/common';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { EditProjectQuotasComponent } from './edit-project-quotas/edit-project-quotas.component';
import { QuotaService } from '../../../../../../ng-swagger-gen/services/quota.service';
import { ProjectService } from '../../../../../../ng-swagger-gen/services/project.service';
import { HttpHeaders, HttpResponse } from '@angular/common/http';

describe('ProjectQuotasComponent', () => {
    let spy: jasmine.Spy;
    let spyUpdate: jasmine.Spy;
    let spyRoute: jasmine.Spy;
    let quotaService: QuotaService;
    let component: ProjectQuotasComponent;
    let fixture: ComponentFixture<ProjectQuotasComponent>;
    let mockQuotaList: Quota[] = [
        {
            id: 1111,
            ref: {
                id: 1111,
                name: 'project1',
                owner_name: 'project1',
            },
            creation_time: '12212112121',
            update_time: '12212112121',
            hard: {
                storage: -1,
            },
            used: {
                storage: 1234,
            },
        },
    ];
    const fakedRouter = {
        navigate() {
            return undefined;
        },
    };
    const fakedErrorHandler = {
        error() {
            return undefined;
        },
        info() {
            return undefined;
        },
    };
    const timeout = (ms: number) => {
        return new Promise(resolve => setTimeout(resolve, ms));
    };
    const fakedProjectService = {
        listProjects() {
            return of([]);
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ProjectQuotasComponent, EditProjectQuotasComponent],
            providers: [
                { provide: ProjectService, useValue: fakedProjectService },
                { provide: ErrorHandler, useValue: fakedErrorHandler },
                { provide: APP_BASE_HREF, useValue: '/' },
                { provide: Router, useValue: fakedRouter },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ProjectQuotasComponent);
        component = fixture.componentInstance;
        component.quotaHardLimitValue = {
            storageLimit: 23,
            storageUnit: 'GB',
        };
        component.loading = true;
        quotaService = fixture.debugElement.injector.get(QuotaService);
        spy = spyOn(quotaService, 'listQuotasResponse').and.callFake(function (
            params: QuotaService.ListQuotasParams
        ): Observable<HttpResponse<Quota[]>> {
            const response: HttpResponse<Array<Quota>> = new HttpResponse<
                Array<Quota>
            >({
                headers: new HttpHeaders({ 'x-total-count': '123' }),
                body: mockQuotaList,
            });
            return of(response).pipe(delay(0));
        });
        spyUpdate = spyOn(quotaService, 'updateQuota').and.returnValue(
            of(null)
        );
        spyRoute = spyOn(
            fixture.debugElement.injector.get(Router),
            'navigate'
        ).and.returnValue(Promise.resolve(true));
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should open edit quota modal', async () => {
        // wait getting list and rendering
        await timeout(10);
        fixture.detectChanges();
        await fixture.whenStable();
        const openEditButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('#open-edit');
        openEditButton.dispatchEvent(new Event('click'));
        fixture.detectChanges();
        await fixture.whenStable();
        const modal: HTMLElement =
            fixture.nativeElement.querySelector('clr-modal');
        expect(modal).toBeTruthy();
    });
    it('should call navigate function', async () => {
        // wait getting list and rendering
        await timeout(10);
        fixture.detectChanges();
        await fixture.whenStable();
        const a: HTMLElement =
            fixture.nativeElement.querySelector('clr-dg-cell a');
        a.dispatchEvent(new Event('click'));
        expect(spyRoute.calls.count()).toEqual(1);
    });
    it('should refresh', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        component.doSearch(null);
        fixture.detectChanges();
        await fixture.whenStable();
        expect(spy.calls.count()).toEqual(2);
    });
    it('should get no quota', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        component.doSearch('test');
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.quotaList.length).toEqual(0);
    });
});
