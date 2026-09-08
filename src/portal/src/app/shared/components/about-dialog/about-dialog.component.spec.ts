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
import { AppConfigService } from '../../../services/app-config.service';
import { SkinableConfig } from '../../../services/skinable-config.service';
import { AboutDialogComponent } from './about-dialog.component';
import { SharedTestingModule } from '../../shared.module';

describe('AboutDialogComponent', () => {
    let component: AboutDialogComponent;
    let fixture: ComponentFixture<AboutDialogComponent>;
    let fakeAppConfigService = {
        getConfig: function () {
            return {
                harbor_version: '1.10',
            };
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
            declarations: [AboutDialogComponent],
            imports: [SharedTestingModule],
            providers: [
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: SkinableConfig, useValue: fakeSkinableConfig },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AboutDialogComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
