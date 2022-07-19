import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { SessionService } from '../../../../shared/services/session.service';
import { of } from 'rxjs';
import { HelmChartDetailComponent } from './chart-detail.component';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('ChartDetailComponent', () => {
    let component: HelmChartDetailComponent;
    let fixture: ComponentFixture<HelmChartDetailComponent>;
    let fakeRouter = null;
    let fakeSessionService = {
        getCurrentUser: function () {
            return { has_admin_role: true };
        },
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [HelmChartDetailComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: ActivatedRoute,
                    useValue: {
                        paramMap: of({ get: key => 'value' }),
                        snapshot: {
                            parent: {
                                data: {
                                    projectResolver: {
                                        role_name: 'admin',
                                    },
                                },
                                params: { id: 1 },
                            },
                            params: {
                                chart: 'chart',
                                version: 1.0,
                            },
                        },
                    },
                },
                { provide: Router, useValue: fakeRouter },
                { provide: SessionService, useValue: fakeSessionService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(HelmChartDetailComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
