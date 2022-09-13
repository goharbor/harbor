import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { of } from 'rxjs';
import { AppConfigService } from '../../../services/app-config.service';
import { SummaryComponent } from './summary.component';
import {
    ProjectService,
    UserPermissionService,
} from '../../../shared/services';
import { ErrorHandler } from '../../../shared/units/error-handler';
import { SessionService } from '../../../shared/services/session.service';
import { EndpointService } from '../../../shared/services/endpoint.service';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('SummaryComponent', () => {
    let component: SummaryComponent;
    let fixture: ComponentFixture<SummaryComponent>;
    let fakeAppConfigService = {
        getConfig() {
            return {
                with_chartmuseum: false,
            };
        },
    };
    let fakeProjectService = {
        getProjectSummary: function () {
            return of();
        },
    };
    let fakeErrorHandler = null;
    let fakeUserPermissionService = {
        hasProjectPermissions: function () {
            return of([true, true]);
        },
    };
    const fakedSessionService = {
        getCurrentUser() {
            return {
                has_admin_role: true,
            };
        },
    };

    const fakedEndpointService = {
        getEndpoint() {
            return of({
                name: 'test',
                url: 'https://test.com',
            });
        },
    };

    const mockedSummaryInformation = {
        repo_count: 0,
        chart_count: 0,
        project_admin_count: 1,
        maintainer_count: 0,
        developer_count: 0,
        registry: {
            name: 'test',
            url: 'https://test.com',
        },
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [SummaryComponent],
            imports: [SharedTestingModule],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            providers: [
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: ProjectService, useValue: fakeProjectService },
                { provide: ErrorHandler, useValue: fakeErrorHandler },
                {
                    provide: UserPermissionService,
                    useValue: fakeUserPermissionService,
                },
                { provide: EndpointService, useValue: fakedEndpointService },
                { provide: SessionService, useValue: fakedSessionService },
                {
                    provide: ActivatedRoute,
                    useValue: {
                        paramMap: of({ get: key => 'value' }),
                        snapshot: {
                            parent: {
                                parent: {
                                    snapshot: {
                                        data: {
                                            projectResolver: { registry_id: 3 },
                                        },
                                    },
                                },
                            },
                        },
                        parent: {
                            parent: {
                                snapshot: {
                                    params: { id: 1 },
                                },
                            },
                        },
                    },
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(SummaryComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should show proxy cache endpoint', async () => {
        component.summaryInformation = mockedSummaryInformation;
        component.isCardView = false;
        fixture.detectChanges();
        await fixture.whenStable();
        const endpoint: HTMLElement =
            fixture.nativeElement.querySelector('#endpoint');
        expect(endpoint).toBeTruthy();
        expect(endpoint.innerText).toEqual('test-https://test.com');
    });

    it('should show card view', async () => {
        component.summaryInformation = mockedSummaryInformation;
        component.isCardView = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const container: HTMLElement =
            fixture.nativeElement.querySelector('.container');
        expect(container).toBeTruthy();
    });

    it('should show two cards', async () => {
        component.summaryInformation = mockedSummaryInformation;
        component.isCardView = true;
        component.hasReadChartPermission = true;
        fixture.detectChanges();
        await fixture.whenStable();
        const cards = fixture.nativeElement.querySelectorAll('.card');
        expect(cards.length).toEqual(2);
    });
});
