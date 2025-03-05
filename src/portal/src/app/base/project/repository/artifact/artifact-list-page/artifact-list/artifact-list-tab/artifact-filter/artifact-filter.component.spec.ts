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
import { ArtifactFilterComponent } from './artifact-filter.component';
import { SharedTestingModule } from '../../../../../../../../shared/shared.module';
import { NO_ERRORS_SCHEMA } from '@angular/core';

describe('ArtifactFilterComponent', () => {
    let component: ArtifactFilterComponent;
    let fixture: ComponentFixture<ArtifactFilterComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [NO_ERRORS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [ArtifactFilterComponent],
        }).compileComponents();

        fixture = TestBed.createComponent(ArtifactFilterComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('Expanding should work', async () => {
        await fixture.whenStable();
        const searchIcon = fixture.nativeElement.querySelector(
            `#${component.searchId}`
        );
        searchIcon.click();
        fixture.detectChanges();
        await fixture.whenStable();
        let selector;
        selector = fixture.nativeElement.querySelector(
            `#${component.typeSelectId}`
        );
        expect(selector).toBeTruthy();
        const searchIconClose = fixture.nativeElement.querySelector(
            `.search-dropdown-toggle`
        );
        searchIconClose.click();
        fixture.detectChanges();
        await fixture.whenStable();
        selector = fixture.nativeElement.querySelector(
            `#${component.typeSelectId}`
        );
        expect(!!selector).toBeFalse();
    });
});
