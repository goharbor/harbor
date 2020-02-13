import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { DependenciesComponent } from "./dependencies.component";
import { ErrorHandler } from '../../../../utils/error-handler';
import { AdditionsService } from '../additions.service';
import { of } from 'rxjs';
import { SERVICE_CONFIG, IServiceConfig } from '../../../../entities/service.config';

describe('DependenciesComponent', () => {
    let component: DependenciesComponent;
    let fixture: ComponentFixture<DependenciesComponent>;
    const mockErrorHandler = {
        error: () => { }
    };
    const mockAdditionsService = {
        getDetailByLink: () => of([])
    };
    const config: IServiceConfig = {
        repositoryBaseEndpoint: "/api/repositories/testing"
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
});
