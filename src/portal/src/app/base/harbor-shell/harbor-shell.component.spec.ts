import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { AppConfigService } from '../..//app-config.service';
import { RouterTestingModule } from '@angular/router/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { SessionService } from '../../shared/session.service';
import { CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { SearchTriggerService } from '../global-search/search-trigger.service';
import { HarborShellComponent } from './harbor-shell.component';
import { ClarityModule } from "@clr/angular";
import { of } from 'rxjs';

describe('HarborShellComponent', () => {
    let component: HarborShellComponent;
    let fixture: ComponentFixture<HarborShellComponent>;
    let fakeSessionService = {
        getCurrentUser: function () {
            return { has_admin_role: true };
        }
    };
    let fakeSearchTriggerService = {
        searchTriggerChan$: {
            subscribe: function () {
            }
        },
        searchCloseChan$: {
            subscribe: function () {
            }
        }
    };
    let fakeAppConfigService = {
        isLdapMode: function () {
            return true;
        },
        isHttpAuthMode: function () {
            return false;
        },
        isOidcMode: function () {
            return false;
        },
        getConfig: function () {
            return {
                with_clair: true
            };
        }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule,
                TranslateModule.forRoot(),
                ClarityModule,
                BrowserAnimationsModule
            ],
            declarations: [HarborShellComponent],
            providers: [
                TranslateService,
                { provide: SessionService, useValue: fakeSessionService },
                { provide: SearchTriggerService, useValue: fakeSearchTriggerService },
                { provide: AppConfigService, useValue: fakeAppConfigService }
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(HarborShellComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
