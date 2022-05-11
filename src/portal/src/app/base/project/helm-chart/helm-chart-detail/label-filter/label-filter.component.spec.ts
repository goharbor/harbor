import { ComponentFixture, TestBed } from '@angular/core/testing';
import { LabelFilterComponent } from './label-filter.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { SharedTestingModule } from '../../../../../shared/shared.module';

describe('LabelFilterComponent', () => {
    let component: LabelFilterComponent;
    let fixture: ComponentFixture<LabelFilterComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [LabelFilterComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(LabelFilterComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
