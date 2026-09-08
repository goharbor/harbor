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
import { of } from 'rxjs';
import { ArtifactInfoComponent } from './artifact-info.component';
import { SharedTestingModule } from 'src/app/shared/shared.module';
import { RepositoryService } from 'ng-swagger-gen/services/repository.service';
import { UserPermissionService } from '../../../../../../../shared/services';

describe('ArtifactInfoComponent', () => {
    let compRepo: ArtifactInfoComponent;
    let fixture: ComponentFixture<ArtifactInfoComponent>;
    let FakedRepositoryService = {
        updateRepository: () => of(null),
        getRepository: () => of({ description: '' }),
    };
    const fakedUserPermissionService = {
        getPermission() {
            return of(true);
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ArtifactInfoComponent],
            providers: [
                {
                    provide: RepositoryService,
                    useValue: FakedRepositoryService,
                },
                {
                    provide: UserPermissionService,
                    useValue: fakedUserPermissionService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactInfoComponent);
        compRepo = fixture.componentInstance;
        fixture.detectChanges();
    });
    it('should create', () => {
        expect(compRepo).toBeTruthy();
    });

    it('should check permission', async () => {
        await fixture.whenStable();
        expect(compRepo.hasEditPermission).toBeTruthy();
    });
});
