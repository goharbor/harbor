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
import { LabelSelectorComponent } from './label-selector.component';
import { SharedTestingModule } from '../../shared.module';
import { LabelService } from '../../../../../ng-swagger-gen/services/label.service';
import { Label } from '../../../../../ng-swagger-gen/models/label';
import { of } from 'rxjs';
import { delay, finalize } from 'rxjs/operators';

describe('LabelSelectorComponent', () => {
    let component: LabelSelectorComponent;
    let fixture: ComponentFixture<LabelSelectorComponent>;
    const mockedLabels: Label[] = [
        {
            id: 1,
            name: 'good',
            scope: 'p',
            project_id: 1,
            color: '#ccc',
        },
        {
            id: 2,
            name: 'bad',
            scope: 'p',
            project_id: 1,
            color: '#ccc',
        },
    ];
    let spy: jasmine.Spy;
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [LabelSelectorComponent],
        }).compileComponents();

        fixture = TestBed.createComponent(LabelSelectorComponent);
        component = fixture.componentInstance;
        spy = spyOn(TestBed.inject(LabelService), 'ListLabels').and.returnValue(
            of(mockedLabels).pipe(delay(0))
        );
        fixture.detectChanges();
        await fixture.whenStable();
        component.loading = false;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render candidates', async () => {
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('hbr-label-piece');
        expect(rows.length).toEqual(2);
    });

    it('owned labels should be checked', async () => {
        await fixture.whenStable();
        component.ownedLabels = [mockedLabels[0]];
        fixture.detectChanges();
        await fixture.whenStable();
        const checkIcon = fixture.nativeElement.querySelector('.check-icon');
        expect(checkIcon.style.visibility).toEqual('hidden');
    });
});
