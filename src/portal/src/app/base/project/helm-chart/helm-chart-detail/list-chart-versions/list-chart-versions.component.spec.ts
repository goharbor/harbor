import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ActivatedRoute, Router } from '@angular/router';
import { SessionService } from '../../../../../shared/services/session.service';
import { ListChartVersionsComponent } from './list-chart-versions.component';

describe('ListChartVersionsComponent', () => {
    let component: ListChartVersionsComponent;
    let fixture: ComponentFixture<ListChartVersionsComponent>;
    let fakeSessionService = {
        getCurrentUser: function () {
            return 'admin';
        },
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [ListChartVersionsComponent],
            imports: [ClarityModule, TranslateModule.forRoot()],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            providers: [
                TranslateService,
                {
                    provide: ActivatedRoute,
                    useValue: {
                        snapshot: {
                            parent: {
                                params: {
                                    id: 1,
                                },
                                data: null,
                            },
                            params: {
                                chart: 'chart',
                            },
                        },
                    },
                },
                { provide: Router, useValue: null },
                { provide: SessionService, useValue: fakeSessionService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ListChartVersionsComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
