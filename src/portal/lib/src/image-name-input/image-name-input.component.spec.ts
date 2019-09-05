import { async, ComponentFixture, TestBed } from "@angular/core/testing";

import { SharedModule } from "../shared/shared.module";
import { ImageNameInputComponent } from "../image-name-input/image-name-input.component";

import { ErrorHandler } from "../error-handler/error-handler";
import { ProjectDefaultService, ProjectService } from "../service/index";
import { ChannelService } from "../channel/index";
import { Project } from "../project-policy-config/project";
import { IServiceConfig, SERVICE_CONFIG } from "../service.config";
import { of } from "rxjs";
import { HttpResponse } from "@angular/common/http";

describe("ImageNameInputComponent (inline template)", () => {
    let comp: ImageNameInputComponent;
    let fixture: ComponentFixture<ImageNameInputComponent>;
    let spy: jasmine.Spy;

    let mockProjects: Project[] = [
        {
            "project_id": 1,
            "name": "project_01",
            "creation_time": "",
        },
        {
            "project_id": 2,
            "name": "project_02",
            "creation_time": "",
        }
    ];

    let config: IServiceConfig = {
        projectBaseEndpoint: "/api/projects/testing"
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                SharedModule
            ],
            declarations: [
                ImageNameInputComponent
            ],
            providers: [
                ErrorHandler,
                ChannelService,
                { provide: SERVICE_CONFIG, useValue: config },
                { provide: ProjectService, useClass: ProjectDefaultService }
            ]
        });
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ImageNameInputComponent);
        comp = fixture.componentInstance;

        let projectService: ProjectService;
        projectService = fixture.debugElement.injector.get(ProjectService);
        spy = spyOn(projectService, "listProjects").and.returnValues(of(new HttpResponse({ body: mockProjects })));
    });

    it("should load data", async(() => {
        expect(spy.calls.any).toBeTruthy();
    }));
});
