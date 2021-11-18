import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ExportCveComponent } from './export-cve.component';
import { SharedTestingModule } from '../../../../../shared/shared.module';

describe('ExportCveComponent', () => {
    let component: ExportCveComponent;
    let fixture: ComponentFixture<ExportCveComponent>;
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ExportCveComponent],
            providers: [],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ExportCveComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
