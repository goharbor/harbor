// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { GlobalSearchService } from './global-search.service';
import { SearchResults } from './search-results';
import { SearchTriggerService } from './search-trigger.service';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { AppConfigService } from '../../../services/app-config.service';
import { ListProjectROComponent } from '../list-project-ro/list-project-ro.component';
import { MessageHandlerService } from '../../services/message-handler.service';
import { SearchResultComponent } from './search-result.component';
import { of } from 'rxjs';
import { AppConfig } from '../../../services/app-config';
import { SharedTestingModule } from '../../shared.module';

describe('SearchResultComponent', () => {
    let component: SearchResultComponent;
    let fixture: ComponentFixture<SearchResultComponent>;
    let fakeSearchResults = null;
    const project = {
        project_id: 1,
        owner_id: 1,
        name: 'library',
        creation_time: Date,
        creation_time_str: 'string',
        deleted: 1,
        owner_name: 'string',
        togglable: true,
        update_time: Date,
        current_user_role_id: 1,
        repo_count: 1,
        chart_count: 1,
        has_project_admin_role: true,
        is_member: true,
        role_name: 'string',
        metadata: {
            public: 'string',
            enable_content_trust: 'string',
            prevent_vul: 'string',
            severity: 'string',
            auto_scan: 'string',
            retention_id: 1,
        },
    };
    let fakeGlobalSearchService = {
        doSearch: () =>
            of({
                project: [project],
                repository: [],
                chart: [],
            }),
    };
    let fakeAppConfigService = {
        getConfig: () => new AppConfig(),
    };
    let searchResult = '';
    let fakeMessageHandlerService = null;
    let fakeSearchTriggerService = {
        searchTriggerChan$: of(searchResult),
        searchCloseChan$: of(null),
        clear: () => null,
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [SearchResultComponent, ListProjectROComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
        })
            .overrideComponent(SearchResultComponent, {
                set: {
                    providers: [
                        {
                            provide: AppConfigService,
                            useValue: fakeAppConfigService,
                        },
                        {
                            provide: MessageHandlerService,
                            useValue: fakeMessageHandlerService,
                        },
                        {
                            provide: SearchTriggerService,
                            useValue: fakeSearchTriggerService,
                        },
                        {
                            provide: GlobalSearchService,
                            useValue: fakeGlobalSearchService,
                        },
                        { provide: SearchResults, useValue: fakeSearchResults },
                    ],
                },
            })
            .compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(SearchResultComponent);
        component = fixture.componentInstance;
        component.stateIndicator = true;
        fixture.autoDetectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should search library', async () => {
        searchResult = 'library';
        component.doSearch(searchResult);
        await fixture.whenStable();
        expect(component.searchResults.project[0].name).toEqual('library');
    });
});
