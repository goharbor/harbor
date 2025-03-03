import { PreferenceSettingsComponent } from './preference-settings.component';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SharedTestingModule } from 'src/app/shared/shared.module';

describe('PreferenceSettingsComponent', () => {
    let component: PreferenceSettingsComponent;
    let fixture: ComponentFixture<PreferenceSettingsComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [PreferenceSettingsComponent],
            imports: [SharedTestingModule],
        }).compileComponents();

        fixture = TestBed.createComponent(PreferenceSettingsComponent);
        component = fixture.componentInstance;

        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
