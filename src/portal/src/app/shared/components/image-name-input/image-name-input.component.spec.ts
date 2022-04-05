import { ComponentFixture, TestBed } from "@angular/core/testing";
import { ImageNameInputComponent } from "./image-name-input.component";
import { ProjectDefaultService, ProjectService } from "../../services";
import { Project } from "../../../base/project/project-config/project-policy-config/project";
import { of } from "rxjs";
import { HttpResponse } from "@angular/common/http";
import { SharedTestingModule } from "../../shared.module";

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
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                SharedTestingModule
            ],
            declarations: [
                ImageNameInputComponent
            ],
            providers: [
                { provide: ProjectService, useClass: ProjectDefaultService }
            ]
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ImageNameInputComponent);
        comp = fixture.componentInstance;

        let projectService: ProjectService;
        projectService = fixture.debugElement.injector.get(ProjectService);
        spy = spyOn(projectService, "listProjects").and.returnValues(of(new HttpResponse({ body: mockProjects })));
    });

    it("should load data", () => {
        expect(spy.calls.any).toBeTruthy();
    });
});
