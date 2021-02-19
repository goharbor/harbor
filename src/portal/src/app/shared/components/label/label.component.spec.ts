import { Label } from "../../services";
import { LabelComponent } from "./label.component";
import { waitForAsync, ComponentFixture, TestBed } from "@angular/core/testing";
import { LabelDefaultService, LabelService } from "../../services";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { FilterComponent } from "../filter/filter.component";
import { ConfirmationDialogComponent } from "../confirmation-dialog";
import { CreateEditLabelComponent } from "./create-edit-label/create-edit-label.component";
import { LabelPieceComponent } from "./label-piece/label-piece.component";
import { InlineAlertComponent } from "../inline-alert/inline-alert.component";
import { ErrorHandler } from "../../units/error-handler";
import { OperationService } from "../operation/operation.service";
import { of } from "rxjs";
import { CURRENT_BASE_HREF } from "../../units/utils";
import { SharedTestingModule } from "../../shared.module";

describe('LabelComponent (inline template)', () => {

    let mockData: Label[] = [
        {
            color: "#9b0d54",
            creation_time: "",
            description: "",
            id: 1,
            name: "label0-g",
            project_id: 0,
            scope: "g",
            update_time: "",
        },
        {
            color: "#9b0d54",
            creation_time: "",
            description: "",
            id: 2,
            name: "label1-g",
            project_id: 0,
            scope: "g",
            update_time: "",
        }
    ];

    let mockOneData: Label = {
        color: "#9b0d54",
        creation_time: "",
        description: "",
        id: 1,
        name: "label0-g",
        project_id: 0,
        scope: "g",
        update_time: "",
    };

    let comp: LabelComponent;
    let fixture: ComponentFixture<LabelComponent>;


    let labelService: LabelService;
    let spy: jasmine.Spy;
    let spyOneLabel: jasmine.Spy;
    beforeEach(waitForAsync(() => {
        TestBed.configureTestingModule({
            imports: [
                SharedTestingModule,
                NoopAnimationsModule
            ],
            declarations: [
                FilterComponent,
                ConfirmationDialogComponent,
                CreateEditLabelComponent,
                LabelComponent,
                LabelPieceComponent,
                InlineAlertComponent
            ],
            providers: [
                ErrorHandler,
                {provide: LabelService, useClass: LabelDefaultService},
                {provide: OperationService}
            ]
        });
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(LabelComponent);
        comp = fixture.componentInstance;

        labelService = fixture.debugElement.injector.get(LabelService);

        spy = spyOn(labelService, 'getLabels').and.returnValues(of(mockData));
        spyOneLabel = spyOn(labelService, 'getLabel').and.returnValues(of(mockOneData));
        fixture.detectChanges();
    });

    it('should retrieve label data', () => {
        fixture.detectChanges();
        expect(spy.calls.any()).toBeTruthy();
    });

    it('should open create label modal', waitForAsync(() => {
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            comp.editLabel([mockOneData]);
            fixture.detectChanges();
            expect(comp.targets[0].name).toEqual('label0-g');
        });
    }));
});
