import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TagFeatureIntegrationComponent } from './tag-feature-integration.component';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('TagFeatureIntegrationComponent', () => {
    let component: TagFeatureIntegrationComponent;
    let fixture: ComponentFixture<TagFeatureIntegrationComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [TagFeatureIntegrationComponent],
            imports: [SharedTestingModule],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TagFeatureIntegrationComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
