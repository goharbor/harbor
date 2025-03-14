// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
