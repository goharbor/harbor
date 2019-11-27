import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ListChartVersionRoComponent } from './list-chart-version-ro.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { of } from 'rxjs';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { SearchTriggerService } from '../../base/global-search/search-trigger.service';
import { ProjectService } from '@harbor/ui';

describe('ListChartVersionRoComponent', () => {
    let component: ListChartVersionRoComponent;
    let fixture: ComponentFixture<ListChartVersionRoComponent>;
    const mockSearchTriggerService = {
        closeSearch: () => { }
    };
    const mockProjectService = {
        listProjects: () => {
            return of(
                {
                    body: []
                }
            );
        }
    };
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
                HttpClientTestingModule
            ],
            declarations: [ListChartVersionRoComponent],
            providers: [
                TranslateService,
                { provide: ProjectService, useValue: mockProjectService },
                { provide: SearchTriggerService, useValue: mockSearchTriggerService }

            ]
        })
            .compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ListChartVersionRoComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
