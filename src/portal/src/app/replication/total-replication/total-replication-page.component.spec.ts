import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { TotalReplicationPageComponent } from './total-replication-page.component';
import {Router, ActivatedRoute} from "@angular/router";
import {ReplicationRule} from "@harbor/ui";
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import {SessionService} from "../../shared/session.service";
import {AppConfigService} from "../../app-config.service";

describe('TotalReplicationPageComponent', () => {
    let component: TotalReplicationPageComponent;
    let fixture: ComponentFixture<TotalReplicationPageComponent>;
    const mockSessionService = {
        getCurrentUser: () => { }
    };
    const mockAppConfigService = {
        getConfig: () => {
            return {
                project_creation_restriction: "",
                with_chartmuseum: ""
            };
        }
    };
    const mockRouter = {
        navigate: () => { }
    };
    const mockActivatedRoute = null;
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
            declarations: [TotalReplicationPageComponent],
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
        fixture = TestBed.createComponent(TotalReplicationPageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
