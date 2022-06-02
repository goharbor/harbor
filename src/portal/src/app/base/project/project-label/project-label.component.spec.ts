import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, ActivatedRoute } from '@angular/router';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { ProjectLabelComponent } from './project-label.component';
import { SessionService } from '../../../shared/services/session.service';
import { UserPermissionService } from '../../../shared/services';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('ProjectLabelComponent', () => {
    let component: ProjectLabelComponent;
    let fixture: ComponentFixture<ProjectLabelComponent>;
    let fakeRouter = null;
    const fakeUserPermissionService = {
        getPermission() {
            return of(true);
        },
    };
    const fakeSessionService = {
        getCurrentUser: function () {
            return { has_admin_role: true };
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ProjectLabelComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            providers: [
                { provide: Router, useValue: fakeRouter },
                {
                    provide: ActivatedRoute,
                    useValue: {
                        snapshot: {
                            parent: {
                                parent: {
                                    params: {
                                        id: 1,
                                    },
                                },
                            },
                        },
                    },
                },
                {
                    provide: UserPermissionService,
                    useValue: fakeUserPermissionService,
                },
                { provide: SessionService, useValue: fakeSessionService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ProjectLabelComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
