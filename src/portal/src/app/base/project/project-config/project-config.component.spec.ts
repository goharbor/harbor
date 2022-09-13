import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { SessionService } from '../../../shared/services/session.service';
import { of } from 'rxjs';
import { ProjectConfigComponent } from './project-config.component';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('ProjectConfigComponent', () => {
    let component: ProjectConfigComponent;
    let fixture: ComponentFixture<ProjectConfigComponent>;
    let fakeSessionService = {
        getCurrentUser: function () {
            return { has_admin_role: true };
        },
    };
    let fakeRouter = null;

    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [ProjectConfigComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: ActivatedRoute,
                    useValue: {
                        paramMap: of({ get: key => 'value' }),
                        snapshot: {
                            parent: {
                                parent: {
                                    params: {
                                        id: 1,
                                        chart: 'chart',
                                        version: 1.0,
                                    },
                                    data: {
                                        projectResolver: {
                                            role_name: 'admin',
                                        },
                                    },
                                },
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
        fixture = TestBed.createComponent(ProjectConfigComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
