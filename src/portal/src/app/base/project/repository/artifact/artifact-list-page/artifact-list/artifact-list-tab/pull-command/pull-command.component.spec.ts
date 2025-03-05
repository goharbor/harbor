import { PullCommandComponent } from './pull-command.component';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SharedTestingModule } from '../../../../../../../../shared/shared.module';
import { ArtifactType } from '../../../../artifact'; // Import the necessary type

describe('PullCommandComponent', () => {
    let component: PullCommandComponent;
    let fixture: ComponentFixture<PullCommandComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [PullCommandComponent],
            imports: [SharedTestingModule],
        }).compileComponents();

        fixture = TestBed.createComponent(PullCommandComponent);
        component = fixture.componentInstance;

        // Mock the artifact input with a valid value
        component.artifact = {
            type: ArtifactType.IMAGE,
            digest: 'sampleDigest',
            tags: [{ name: 'latest' }],
        };

        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
