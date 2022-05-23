import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ClearingJobComponent } from './clearing-job.component';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('GcPageComponent', () => {
    let component: ClearingJobComponent;
    let fixture: ComponentFixture<ClearingJobComponent>;
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ClearingJobComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ClearingJobComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
