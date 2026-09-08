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
import { ExportCveComponent } from './export-cve.component';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { Router } from '@angular/router';

describe('ExportCveComponent', () => {
    let component: ExportCveComponent;
    let fixture: ComponentFixture<ExportCveComponent>;
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ExportCveComponent],
            providers: [],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ExportCveComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('goToLabels should close the modal and navigate to /harbor/labels', () => {
        const router = TestBed.inject(Router);
        const navigateSpy = spyOn(router, 'navigate');
        const closeSpy = spyOn(component, 'close').and.callThrough();

        component.opened = true;
        component.goToLabels();

        expect(closeSpy).toHaveBeenCalled();
        expect(component.opened).toBeFalse();
        expect(navigateSpy).toHaveBeenCalledOnceWith(['/harbor/labels']);
    });
});
