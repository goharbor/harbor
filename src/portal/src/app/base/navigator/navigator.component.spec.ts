import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { SessionService } from '../../shared/session.service';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { PlatformLocation } from '@angular/common';
import { NavigatorComponent } from './navigator.component';
import { RouterTestingModule } from '@angular/router/testing';
import { CookieService } from 'ngx-cookie';
import { AppConfigService } from '../../app-config.service';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { SearchTriggerService } from '../global-search/search-trigger.service';
import { SkinableConfig } from "../../skinable-config.service";

describe('NavigatorComponent', () => {
    let component: NavigatorComponent;
    let fixture: ComponentFixture<NavigatorComponent>;
    let fakeSessionService = {
        getCurrentUser: function () {
            return {
                username: 'abc',
                has_admin_role: true
            };
        }
    };
    let fakePlatformLocation = null;
    let fakeCookieService = null;
    let fakeAppConfigService = {
        isIntegrationMode: function () {
            return true;
        },
        getConfig: function () {
            return {
                has_ca_root: true,
                read_only: false
            };
        },
        getAdmiralEndpoint: function () {

        }
    };
    let fakeMessageHandlerService = null;
    let fakeSearchTriggerService = null;
    let fakeSkinableConfig = {
        getSkinConfig: function () {
            return { projects: "abc" };
        }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                RouterTestingModule
            ],
            declarations: [NavigatorComponent],
            providers: [
                TranslateService,
                { provide: SessionService, useValue: fakeSessionService },
                { provide: PlatformLocation, useValue: fakePlatformLocation },
                { provide: CookieService, useValue: fakeCookieService },
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: MessageHandlerService, useValue: fakeMessageHandlerService },
                { provide: SearchTriggerService, useValue: fakeSearchTriggerService },
                { provide: SkinableConfig, useValue: fakeSkinableConfig }
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(NavigatorComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
