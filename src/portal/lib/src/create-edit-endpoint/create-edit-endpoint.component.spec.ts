import {
  ComponentFixture,
  TestBed,
  async
} from "@angular/core/testing";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { SharedModule } from "../shared/shared.module";

import { FilterComponent } from "../filter/filter.component";

import { CreateEditEndpointComponent } from "../create-edit-endpoint/create-edit-endpoint.component";
import { InlineAlertComponent } from "../inline-alert/inline-alert.component";
import { ErrorHandler } from "../error-handler/error-handler";
import { Endpoint } from "../service/interface";
import {
  EndpointService,
  EndpointDefaultService
} from "../service/endpoint.service";
import { IServiceConfig, SERVICE_CONFIG } from "../service.config";
import { of } from "rxjs";
describe("CreateEditEndpointComponent (inline template)", () => {
  let mockData: Endpoint = {
    id: 1,
    credential: {
      access_key: "admin",
      access_secret: "",
      type: "basic"
    },
    description: "test",
    insecure: false,
    name: "target_01",
    type: "Harbor",
    url: "https://10.117.4.151"
  };

  let mockAdapters = ['harbor', 'docker hub'];

  let comp: CreateEditEndpointComponent;
  let fixture: ComponentFixture<CreateEditEndpointComponent>;

  let config: IServiceConfig = {
    systemInfoEndpoint: "/api/endpoints/testing"
  };

  let endpointService: EndpointService;

  let spy: jasmine.Spy;
  let spyAdapter: jasmine.Spy;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [SharedModule, NoopAnimationsModule],
      declarations: [
        FilterComponent,
        CreateEditEndpointComponent,
        InlineAlertComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: EndpointService, useClass: EndpointDefaultService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(CreateEditEndpointComponent);
    comp = fixture.componentInstance;

    endpointService = fixture.debugElement.injector.get(EndpointService);
    spyAdapter = spyOn(endpointService, "getAdapters").and.returnValue(
      of(mockAdapters)
    );

    spy = spyOn(endpointService, "getEndpoint").and.returnValue(
      of(mockData)
    );
    fixture.detectChanges();

    comp.openCreateEditTarget(true, 1);
    fixture.detectChanges();
  });

  it("should be created", () => {
    fixture.detectChanges();
    expect(comp).toBeTruthy();
  });

  it("should get endpoint be called", async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      expect(spy.calls.any()).toBeTruthy();
    });
  }));

  it("should get endpoint and open modal", async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      expect(comp.target.name).toEqual("target_01");
    });
  }));

  it("should endpoint be initialized", () => {
    fixture.detectChanges();
    expect(config.systemInfoEndpoint).toEqual("/api/endpoints/testing");
  });
});
