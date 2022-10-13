import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SharedTestingModule } from 'src/app/shared/shared.module';
import { JobServiceDashboardComponent } from './job-service-dashboard.component';
import { NO_ERRORS_SCHEMA } from '@angular/core';

describe('JobServiceDashboardComponent', () => {
    let component: JobServiceDashboardComponent;
    let fixture: ComponentFixture<JobServiceDashboardComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [NO_ERRORS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [JobServiceDashboardComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(JobServiceDashboardComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
