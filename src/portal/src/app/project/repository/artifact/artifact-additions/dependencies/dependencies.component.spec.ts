import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { DependenciesComponent } from "./dependencies.component";
import { AdditionsService } from '../additions.service';
import { of } from 'rxjs';
import { ArtifactDependency } from "../models";
import { AdditionLink } from "../../../../../../../ng-swagger-gen/models/addition-link";
import { IServiceConfig, SERVICE_CONFIG } from "../../../../../../lib/entities/service.config";
import { ErrorHandler } from "../../../../../../lib/utils/error-handler";
import { CURRENT_BASE_HREF } from "../../../../../../lib/utils/utils";


describe('DependenciesComponent', () => {
    let component: DependenciesComponent;
    let fixture: ComponentFixture<DependenciesComponent>;
    const mockErrorHandler = {
        error: () => { }
    };
    const mockedDependencies: ArtifactDependency[] = [
        {
            name: 'abc',
            version: 'v1.0',
            repository: 'test1'
        },
        {
            name: 'def',
            version: 'v1.1',
            repository: 'test2'
        }
    ];
    const mockAdditionsService = {
        getDetailByLink: () => of(mockedDependencies)
    };
    const mockedLink: AdditionLink = {
        absolute: false,
        href: '/test'
    };
    const config: IServiceConfig = {
        repositoryBaseEndpoint: CURRENT_BASE_HREF + "/repositories/testing"
    };
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot()
            ],
            declarations: [DependenciesComponent],
            providers: [
                TranslateService,
                { provide: SERVICE_CONFIG, useValue: config },

                {
                    provide: ErrorHandler, useValue: mockErrorHandler
                },
                { provide: AdditionsService, useValue: mockAdditionsService },
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(DependenciesComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should get dependencies and render', async () => {
        component.dependenciesLink = mockedLink;
        component.ngOnInit();
        fixture.detectChanges();
        await fixture.whenStable();
        const trs = fixture.nativeElement.getElementsByTagName('tr');
        expect(trs.length).toEqual(3);
    });
});
