import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { TagRepositoryComponent } from './tag-repository.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { of } from 'rxjs';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ActivatedRoute, Router } from '@angular/router';

import { AppConfigService } from '../../app-config.service';
import { SessionService } from '../../shared/session.service';
describe('TagRepositoryComponent', () => {
    let component: TagRepositoryComponent;
    let fixture: ComponentFixture<TagRepositoryComponent>;
    const mockSessionService = {
        getCurrentUser: () => { }
    };
    const mockAppConfigService = {
        getConfig: () => {
            return {
                project_creation_restriction: "",
                with_chartmuseum: "",
                with_notary: "",
                with_clair: "",
                with_admiral: "",
                registry_url: "",
            };
        }
    };
    const mockRouter = {
        navigate: () => { }
    };
    const mockActivatedRoute = {
        RouterparamMap: of({ get: (key) => 'value' }),
        snapshot: {
            params: { id: 1 },
            parent: {
                params: { id: 1 },

            },
            data: {
                projectResolver: {
                    has_project_admin_role: true,
                    current_user_role_id: 3,
                }
            }
        },
        data: of({
            projectResolver: {
                ismember: true,
                role_name: 'master',
            }
        })
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
            declarations: [TagRepositoryComponent],
            providers: [
                TranslateService,
                { provide: SessionService, useValue: mockSessionService },
                { provide: AppConfigService, useValue: mockAppConfigService },
                { provide: Router, useValue: mockRouter },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(TagRepositoryComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
