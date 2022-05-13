import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ChartDetailDependencyComponent } from './chart-detail-dependency.component';
import { SharedTestingModule } from '../../../../../shared/shared.module';

describe('ChartDetailDependencyComponent', () => {
    let component: ChartDetailDependencyComponent;
    let fixture: ComponentFixture<ChartDetailDependencyComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ChartDetailDependencyComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ChartDetailDependencyComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
