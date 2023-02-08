import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SessionService } from '../../services/session.service';
import { Component, NO_ERRORS_SCHEMA } from '@angular/core';
import { PlatformLocation } from '@angular/common';
import { NavigatorComponent } from './navigator.component';
import { CookieService } from 'ngx-cookie';
import { AppConfigService } from '../../../services/app-config.service';
import { MessageHandlerService } from '../../services/message-handler.service';
import { SearchTriggerService } from '../global-search/search-trigger.service';
import { SkinableConfig } from '../../../services/skinable-config.service';
import { SharedTestingModule } from '../../shared.module';

describe('NavigatorComponent', () => {
    let component: TestComponentWrapperComponent;
    let fixture: ComponentFixture<TestComponentWrapperComponent>;
    let fakeSessionService = {
        getCurrentUser: function () {
            return {
                username: 'abc',
                has_admin_role: true,
            };
        },
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
                read_only: false,
            };
        },
        getAdmiralEndpoint: function () {},
    };
    let fakeMessageHandlerService = null;
    let fakeSkinableConfig = {
        getSkinConfig: function () {
            return { projects: 'abc' };
        },
    };
    let fakeSearchTriggerService = {
        searchClearChan$: {
            subscribe: function () {},
        },
        triggerSearch() {
            return undefined;
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [NavigatorComponent, TestComponentWrapperComponent],
            providers: [
                { provide: SessionService, useValue: fakeSessionService },
                { provide: PlatformLocation, useValue: fakePlatformLocation },
                { provide: CookieService, useValue: fakeCookieService },
                { provide: AppConfigService, useValue: fakeAppConfigService },
                {
                    provide: MessageHandlerService,
                    useValue: fakeMessageHandlerService,
                },
                {
                    provide: SearchTriggerService,
                    useValue: fakeSearchTriggerService,
                },
                { provide: SkinableConfig, useValue: fakeSkinableConfig },
            ],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TestComponentWrapperComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});

// clr-header should only be used inside of a clr-main-container
@Component({
    selector: 'test-component-wrapper',
    template:
        '<clr-main-container><navigator></navigator></clr-main-container>',
})
class TestComponentWrapperComponent {}
