import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ReplicationPageComponent } from './replication-page.component';
import { ActivatedRoute, Router } from '@angular/router';

import { SessionService } from "../shared/session.service";
import { Project } from "../project/project";
import { ReplicationComponent, UserPermissionService, USERSTATICPERMISSION, ErrorHandler, ProjectService } from "@harbor/ui";
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { of } from 'rxjs';
describe('ReplicationPageComponent', () => {
    let component: ReplicationPageComponent;
    let fixture: ComponentFixture<ReplicationPageComponent>;
    const mockSessionService = {
        getCurrentUser: () => { }
    };
    const mockUserPermissionService = {
        getPermission() {
            return of(true);
        }
    };
    const mockErrorHandler = {
        error: () => { }
    };
    const mockProjectService = {
        listProjects: () => {
            return of({
                body: []
            });
        }
    };
    const mockActivatedRoute = {
        RouterparamMap: of({ get: (key) => 'value' }),
        snapshot: {
            parent: {
                params: { id: 1 },
                data: {
                    projectResolver: {
                        ismember: true,
                        name: 'library',
                    }
                }
            },
            queryParams: {
                is_create: ""
            }
        }
    };
    const mockRouter = {
        navigate: () => { }
    };
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
                HttpClientTestingModule
            ],
            declarations: [ReplicationPageComponent],
            providers: [
                TranslateService,
                { provide: ErrorHandler, useValue: mockErrorHandler },
                { provide: SessionService, useValue: mockSessionService },
                { provide: UserPermissionService, useValue: mockUserPermissionService },
                { provide: ProjectService, useValue: mockProjectService },
                { provide: ErrorHandler, useValue: mockErrorHandler },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
                { provide: Router, useValue: mockRouter },

            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ReplicationPageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
