import { ComponentFixture, TestBed } from '@angular/core/testing';

import { CopyArtifactComponent } from './copy-artifact.component';
import { SharedTestingModule } from '../../../../../../../../shared/shared.module';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { RouterTestingModule } from '@angular/router/testing';
import { ArtifactService } from 'ng-swagger-gen/services/artifact.service';
import { of } from 'rxjs';

describe('CopyArtifactComponent', () => {
    let component: CopyArtifactComponent;
    let fixture: ComponentFixture<CopyArtifactComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [NO_ERRORS_SCHEMA],
            declarations: [CopyArtifactComponent],
            imports: [SharedTestingModule, RouterTestingModule],
        }).compileComponents();

        fixture = TestBed.createComponent(CopyArtifactComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should open modal', async () => {
        component.retag(`sha256@test`);
        fixture.detectChanges();
        await fixture.whenStable();
        const modal = fixture.nativeElement.querySelector(`clr-modal`);
        expect(modal).toBeTruthy();
    });
    it('should call retag API', async () => {
        const artifactService: ArtifactService =
            TestBed.inject(ArtifactService);
        const spy: jasmine.Spy = spyOn(
            artifactService,
            `CopyArtifact`
        ).and.returnValue(of(null));
        component.retagDialogOpened = true;
        component.imageNameInput.imageNameForm.reset({
            repoName: `test`,
            projectName: `test`,
        });
        fixture.detectChanges();
        await fixture.whenStable();
        const btn: HTMLButtonElement =
            fixture.nativeElement.querySelector(`.btn-primary`);
        btn.click();
        fixture.detectChanges();
        expect(spy.calls.count()).toEqual(1);
    });
});
