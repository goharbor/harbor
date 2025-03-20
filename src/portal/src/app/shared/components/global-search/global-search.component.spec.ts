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
import {
    ComponentFixture,
    fakeAsync,
    TestBed,
    tick,
} from '@angular/core/testing';
import { GlobalSearchComponent } from './global-search.component';
import { SearchTriggerService } from './search-trigger.service';
import { AppConfigService } from '../../../services/app-config.service';
import { SkinableConfig } from '../../../services/skinable-config.service';
import { SharedTestingModule } from '../../shared.module';

describe('GlobalSearchComponent', () => {
    let component: GlobalSearchComponent;
    let fixture: ComponentFixture<GlobalSearchComponent>;
    let fakeSearchTriggerService = {
        searchClearChan$: {
            subscribe: function () {},
        },
        triggerSearch() {
            return undefined;
        },
    };
    let fakeAppConfigService = {
        isIntegrationMode: function () {
            return true;
        },
    };
    let fakeSkinableConfig = {
        getSkinConfig: function () {
            return {
                headerBgColor: {
                    darkMode: '',
                    lightMode: '',
                },
                loginBgImg: '',
                loginTitle: '',
                product: {
                    name: '',
                    logo: '',
                    introduction: '',
                },
            };
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [GlobalSearchComponent],
            providers: [
                {
                    provide: SearchTriggerService,
                    useValue: fakeSearchTriggerService,
                },
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: SkinableConfig, useValue: fakeSkinableConfig },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(GlobalSearchComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should trigger search', fakeAsync(async () => {
        const service: SearchTriggerService = TestBed.get(SearchTriggerService);
        const spy: jasmine.Spy = spyOn(
            service,
            'triggerSearch'
        ).and.callThrough();
        const input: HTMLInputElement =
            fixture.nativeElement.querySelector('#search_input');
        expect(input).toBeTruthy();
        input.value = 'test';
        input.dispatchEvent(new Event('keyup'));
        tick(500);
        fixture.detectChanges();
        await fixture.whenStable();
        expect(spy.calls.count()).toEqual(1);
    }));
});
