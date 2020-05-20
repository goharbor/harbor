import {Label} from "../../services/interface";
import {LabelComponent} from "./label.component";
import {async, ComponentFixture, TestBed} from "@angular/core/testing";
import {LabelDefaultService, LabelService} from "../../services/label.service";
import {SharedModule} from "../../utils/shared/shared.module";
import {NoopAnimationsModule} from "@angular/platform-browser/animations";
import {FilterComponent} from "../filter/filter.component";
import {ConfirmationDialogComponent} from "../confirmation-dialog/confirmation-dialog.component";
import {CreateEditLabelComponent} from "../create-edit-label/create-edit-label.component";
import {LabelPieceComponent} from "../label-piece/label-piece.component";
import {InlineAlertComponent} from "../inline-alert/inline-alert.component";
import {ErrorHandler} from "../../utils/error-handler/error-handler";

import {IServiceConfig, SERVICE_CONFIG} from "../../entities/service.config";
import { OperationService } from "../operation/operation.service";
import { of } from "rxjs";
import { CURRENT_BASE_HREF } from "../../utils/utils";

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

    let config: IServiceConfig = {
        systemInfoEndpoint: CURRENT_BASE_HREF + '/label/testing'
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                SharedModule,
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
                { provide: SERVICE_CONFIG, useValue: config },
                {provide: LabelService, useClass: LabelDefaultService},
                { provide: OperationService }
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

    it('should open create label modal', async(() => {
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            comp.editLabel([mockOneData]);
            fixture.detectChanges();
            expect(comp.targets[0].name).toEqual('label0-g');
        });
    }));

    /*it('should open to edit existing label', async() => {
        fixture.detectChanges();
        fixture.whenStable().then(() => {
        let de: DebugElement = fixture.debugElement.query(del => del.classes['active']);
        expect(de).toBeTruthy();
        fixture.detectChanges();
        click(de);
        fixture.detectChanges();

        let deInput: DebugElement = fixture.debugElement.query(By.css['input']);
        expect(deInput).toBeTruthy();
        let elInput: HTMLElement = deInput.nativeElement;
        expect(elInput).toBeTruthy();
        expect(elInput.textContent).toEqual('label1-g');

        })
    })*/

});
