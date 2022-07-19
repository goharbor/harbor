import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ListChartVersionRoComponent } from './list-chart-version-ro.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { SearchTriggerService } from '../global-search/search-trigger.service';
import { ProjectService } from '../../services';
import { SharedTestingModule } from '../../shared.module';

describe('ListChartVersionRoComponent', () => {
    let component: ListChartVersionRoComponent;
    let fixture: ComponentFixture<ListChartVersionRoComponent>;
    const mockSearchTriggerService = {
        closeSearch: () => {},
    };
    const mockProjectService = {
        listProjects: () => {
            return of({
                body: [],
            });
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [ListChartVersionRoComponent],
            providers: [
                { provide: ProjectService, useValue: mockProjectService },
                {
                    provide: SearchTriggerService,
                    useValue: mockSearchTriggerService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ListChartVersionRoComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
