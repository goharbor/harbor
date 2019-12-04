import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { GlobalSearchService } from './global-search.service';
import { SearchResults } from './search-results';
import { SearchTriggerService } from './search-trigger.service';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { AppConfigService } from './../../app-config.service';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { SearchResultComponent } from './search-result.component';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('SearchResultComponent', () => {
    let component: SearchResultComponent;
    let fixture: ComponentFixture<SearchResultComponent>;
    let fakeSearchResults = null;
    let fakeGlobalSearchService = null;
    let fakeAppConfigService = null;
    let fakeMessageHandlerService = null;
    let fakeSearchTriggerService = {
        searchTriggerChan$: {
            pipe() {
                return {
                   subscribe() {
                   }
                };
            }
        },
        searchCloseChan$: {
            subscribe() {
            }
        }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                HttpClientTestingModule
            ],
            declarations: [SearchResultComponent],
            providers: [
                TranslateService,
                { provide: GlobalSearchService, useValue: fakeGlobalSearchService },
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: MessageHandlerService, useValue: fakeMessageHandlerService },
                { provide: SearchTriggerService, useValue: fakeSearchTriggerService },
                { provide: SearchResults, fakeSearchResults }

            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(SearchResultComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
