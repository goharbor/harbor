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
