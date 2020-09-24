import { ComponentFixture, TestBed, async } from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { SharedModule } from "../../utils/shared/shared.module";
import { FilterComponent } from "../filter/filter.component";
import { InlineAlertComponent } from "../inline-alert/inline-alert.component";
import { ErrorHandler } from "../../utils/error-handler/error-handler";
import { Label } from "../../services/interface";
import { IServiceConfig, SERVICE_CONFIG } from "../../entities/service.config";
import { CreateEditLabelComponent } from "./create-edit-label.component";
import { LabelDefaultService, LabelService } from "../../services/label.service";
import { of } from "rxjs";
import { CURRENT_BASE_HREF } from "../../utils/utils";

describe("CreateEditLabelComponent (inline template)", () => {
  let mockOneData: Label = {
    color: "#9b0d54",
    creation_time: "",
    description: "",
    id: 1,
    name: "label0-g",
    project_id: 0,
    scope: "g",
    update_time: ""
  };

  let comp: CreateEditLabelComponent;
  let fixture: ComponentFixture<CreateEditLabelComponent>;

  let config: IServiceConfig = {
    systemInfoEndpoint: CURRENT_BASE_HREF + "/label/testing"
  };

  let labelService: LabelService;

  let spy: jasmine.Spy;
  let spyOne: jasmine.Spy;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [SharedModule, NoopAnimationsModule],
      declarations: [
        FilterComponent,
        CreateEditLabelComponent,
        InlineAlertComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: LabelService, useClass: LabelDefaultService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(CreateEditLabelComponent);
    comp = fixture.componentInstance;

    labelService = fixture.debugElement.injector.get(LabelService);

    spy = spyOn(labelService, "getLabels").and.returnValue(
      of([mockOneData])
    );
    spyOne = spyOn(labelService, "createLabel").and.returnValue(
      of(mockOneData)
    );

    fixture.detectChanges();

    comp.openModal();
    fixture.detectChanges();
  });

  it("should be created", () => {
    fixture.detectChanges();
    expect(comp).toBeTruthy();
  });

  it("should get label and open modal", async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      expect(comp.labelModel.name).toEqual("");
    });
  }));
});
