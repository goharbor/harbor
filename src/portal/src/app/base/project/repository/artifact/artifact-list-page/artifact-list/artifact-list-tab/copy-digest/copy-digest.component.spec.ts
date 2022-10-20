import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CopyDigestComponent } from './copy-digest.component';
import { SharedTestingModule } from '../../../../../../../../shared/shared.module';

describe('CopyDigestComponent', () => {
    let component: CopyDigestComponent;
    let fixture: ComponentFixture<CopyDigestComponent>;
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [CopyDigestComponent],
            imports: [SharedTestingModule],
        }).compileComponents();

        fixture = TestBed.createComponent(CopyDigestComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should show right digest', async () => {
        const digest: string = 'sha256@test';
        component.showDigestId(digest);
        fixture.detectChanges();
        await fixture.whenStable();
        const textArea: HTMLTextAreaElement =
            fixture.nativeElement.querySelector(`textarea`);
        expect(textArea.textContent).toEqual(digest);
    });
});
