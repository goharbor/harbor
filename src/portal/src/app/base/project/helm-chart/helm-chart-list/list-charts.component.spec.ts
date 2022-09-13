import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { SessionService } from '../../../../shared/services/session.service';
import { ListChartsComponent } from './list-charts.component';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('ListChartsComponent', () => {
    let component: ListChartsComponent;
    let fixture: ComponentFixture<ListChartsComponent>;
    let fakeSessionService = {
        getCurrentUser: function () {
            return 'admin';
        },
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [ListChartsComponent],
            imports: [SharedTestingModule],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            providers: [
                {
                    provide: ActivatedRoute,
                    useValue: {
                        snapshot: {
                            parent: {
                                parent: {
                                    params: {
                                        id: 1,
                                    },
                                    data: null,
                                },
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
        fixture = TestBed.createComponent(ListChartsComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
