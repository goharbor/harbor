import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, ActivatedRoute } from '@angular/router';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { ProjectLabelComponent } from './project-label.component';
import { SessionService } from '../../shared/session.service';
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { UserPermissionService } from "../../../lib/services";

describe('ProjectLabelComponent', () => {
    let component: ProjectLabelComponent;
    let fixture: ComponentFixture<ProjectLabelComponent>;
    let fakeRouter = null;
    const fakeUserPermissionService = {
        getPermission() {
            return of(true);
        }
    };
    const fakeSessionService = {
        getCurrentUser: function () {
            return { has_admin_role: true };
        }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ProjectLabelComponent],
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            providers: [
                ErrorHandler,
                { provide: Router, useValue: fakeRouter },
                {
                    provide: ActivatedRoute, useValue: {
                        snapshot: {
                            parent: {
                                params: {
                                    id: 1
                                }
                            }
                        }
                    }
                },
                { provide: UserPermissionService, useValue: fakeUserPermissionService },
                { provide: SessionService, useValue: fakeSessionService }
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ProjectLabelComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
