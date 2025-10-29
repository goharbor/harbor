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
import { SharedTestingModule } from 'src/app/shared/shared.module';
import { SubAccessoriesComponent } from './sub-accessories.component';
import { Accessory } from '../../../../../../../../../../ng-swagger-gen/models/accessory';
import { AccessoryType } from '../../../../artifact';
import { ArtifactService as NewArtifactService } from '../../../../../../../../../../ng-swagger-gen/services/artifact.service';
import { of } from 'rxjs';
import {
    ArtifactDefaultService,
    ArtifactService,
} from '../../../../artifact.service';
import { delay } from 'rxjs';

describe('SubAccessoriesComponent', () => {
    const mockedAccessories: Accessory[] = [
        {
            id: 1,
            artifact_id: 1,
            digest: 'sha256:test',
            type: AccessoryType.COSIGN,
            size: 1024,
        },
        {
            id: 2,
            artifact_id: 2,
            digest: 'sha256:test2',
            type: AccessoryType.COSIGN,
            size: 1024,
        },
        {
            id: 3,
            artifact_id: 3,
            digest: 'sha256:test3',
            type: AccessoryType.COSIGN,
            size: 1024,
        },
        {
            id: 4,
            artifact_id: 4,
            digest: 'sha256:test4',
            type: AccessoryType.COSIGN,
            size: 1024,
        },
        {
            id: 5,
            artifact_id: 5,
            digest: 'sha256:test5',
            type: AccessoryType.COSIGN,
            size: 1024,
        },
    ];

    const page2: Accessory[] = [
        {
            id: 6,
            artifact_id: 6,
            digest: 'sha256:test6',
            type: AccessoryType.COSIGN,
            size: 1024,
        },
    ];

    const mockedArtifactService = {
        listAccessories() {
            return of(page2).pipe(delay(0));
        },
        listAccessoriesResponse() {
            return of({}).pipe(delay(0));
        },
    };

    let component: SubAccessoriesComponent;
    let fixture: ComponentFixture<SubAccessoriesComponent>;
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [SubAccessoriesComponent],
            providers: [
                {
                    provide: NewArtifactService,
                    useValue: mockedArtifactService,
                },
                { provide: ArtifactService, useClass: ArtifactDefaultService },
            ],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(SubAccessoriesComponent);
        component = fixture.componentInstance;
        component.accessories = mockedAccessories;
        component.total = 6;
        fixture.autoDetectChanges();
    });
    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render rows', async () => {
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(5);
    });

    it('should render next page', async () => {
        await fixture.whenStable();
        const nextPageButton: HTMLButtonElement =
            fixture.nativeElement.querySelector('.pagination-next');
        nextPageButton.click();
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(1);
    });
});
