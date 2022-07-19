import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SharedTestingModule } from '../../shared.module';
import { FilterComponent } from './filter.component';

describe('FilterComponent', () => {
    let component: FilterComponent;
    let fixture: ComponentFixture<FilterComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [FilterComponent],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(FilterComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should be focused', async () => {
        await fixture.whenStable();
        const searchIcon: HTMLElement =
            fixture.nativeElement.querySelector('.search-btn');
        searchIcon.click();
        fixture.detectChanges();
        await fixture.whenStable();
        const input: HTMLInputElement = fixture.nativeElement.querySelector(
            '.filter-input:focus'
        );
        expect(input).toBeTruthy();
    });
});
