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
import { Label } from '../../services';
import { LabelComponent } from './label.component';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { FilterComponent } from '../filter/filter.component';
import { ConfirmationDialogComponent } from '../confirmation-dialog';
import { CreateEditLabelComponent } from './create-edit-label/create-edit-label.component';
import { LabelPieceComponent } from './label-piece/label-piece.component';
import { InlineAlertComponent } from '../inline-alert/inline-alert.component';
import { OperationService } from '../operation/operation.service';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../shared.module';
import { LabelService } from '../../../../../ng-swagger-gen/services/label.service';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Registry } from '../../../../../ng-swagger-gen/models/registry';

describe('LabelComponent (inline template)', () => {
    let mockData: Label[] = [
        {
            color: '#9b0d54',
            creation_time: '',
            description: '',
            id: 1,
            name: 'label0-g',
            project_id: 0,
            scope: 'g',
            update_time: '',
        },
        {
            color: '#9b0d54',
            creation_time: '',
            description: '',
            id: 2,
            name: 'label1-g',
            project_id: 0,
            scope: 'g',
            update_time: '',
        },
    ];

    let mockOneData: Label = {
        color: '#9b0d54',
        creation_time: '',
        description: '',
        id: 1,
        name: 'label0-g',
        project_id: 0,
        scope: 'g',
        update_time: '',
    };

    let comp: LabelComponent;
    let fixture: ComponentFixture<LabelComponent>;

    let labelService: LabelService;
    let spy: jasmine.Spy;
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [
                FilterComponent,
                ConfirmationDialogComponent,
                CreateEditLabelComponent,
                LabelComponent,
                LabelPieceComponent,
                InlineAlertComponent,
            ],
            providers: [{ provide: OperationService }],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(LabelComponent);
        comp = fixture.componentInstance;

        labelService = fixture.debugElement.injector.get(LabelService);
        const response: HttpResponse<Array<Registry>> = new HttpResponse<
            Array<Registry>
        >({
            headers: new HttpHeaders({ 'x-total-count': [].length.toString() }),
            body: mockData,
        });
        spy = spyOn(labelService, 'ListLabelsResponse').and.returnValues(
            of(response).pipe(delay(0))
        );
        fixture.detectChanges();
    });

    it('should retrieve label data', () => {
        fixture.detectChanges();
        expect(spy.calls.any()).toBeTruthy();
    });

    it('should open create label modal', () => {
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            comp.editLabel([mockOneData]);
            fixture.detectChanges();
            expect(comp.targets[0].name).toEqual('label0-g');
        });
    });
});
