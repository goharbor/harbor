import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { FilterComponent } from '../../filter/filter.component';
import { InlineAlertComponent } from '../../inline-alert/inline-alert.component';
import { ErrorHandler } from '../../../units/error-handler';
import { Label } from '../../../services';
import { CreateEditLabelComponent } from './create-edit-label.component';
import { LabelDefaultService, LabelService } from '../../../services';
import { of } from 'rxjs';
import { SharedTestingModule } from '../../../shared.module';

describe('CreateEditLabelComponent (inline template)', () => {
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

    let comp: CreateEditLabelComponent;
    let fixture: ComponentFixture<CreateEditLabelComponent>;
    let labelService: LabelService;

    let spy: jasmine.Spy;
    let spyOne: jasmine.Spy;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule, NoopAnimationsModule],
            declarations: [
                FilterComponent,
                CreateEditLabelComponent,
                InlineAlertComponent,
            ],
            providers: [
                ErrorHandler,
                { provide: LabelService, useClass: LabelDefaultService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(CreateEditLabelComponent);
        comp = fixture.componentInstance;

        labelService = fixture.debugElement.injector.get(LabelService);

        spy = spyOn(labelService, 'getLabels').and.returnValue(
            of([mockOneData])
        );
        spyOne = spyOn(labelService, 'createLabel').and.returnValue(
            of(mockOneData)
        );

        fixture.detectChanges();

        comp.openModal();
        fixture.detectChanges();
    });

    it('should be created', () => {
        fixture.detectChanges();
        expect(comp).toBeTruthy();
    });

    it('should get label and open modal', () => {
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            expect(comp.labelModel.name).toEqual('');
        });
    });
});
