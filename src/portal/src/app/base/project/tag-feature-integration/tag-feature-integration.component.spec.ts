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
import { TagFeatureIntegrationComponent } from './tag-feature-integration.component';
import { SharedTestingModule } from '../../../shared/shared.module';
import { UserPermissionService } from '../../../shared/services';
import { ActivatedRoute } from '@angular/router';
import { of } from 'rxjs';

describe('TagFeatureIntegrationComponent', () => {
    let component: TagFeatureIntegrationComponent;
    let fixture: ComponentFixture<TagFeatureIntegrationComponent>;

    const mockActivatedRoute = {
        snapshot: {
            parent: {
                parent: {
                    params: { id: 1 },
                },
            },
        },
    };
    const mockUserPermissionService = {
        getPermission() {
            return of(true);
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [TagFeatureIntegrationComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: UserPermissionService,
                    useValue: mockUserPermissionService,
                },
                {
                    provide: ActivatedRoute,
                    useValue: mockActivatedRoute,
                },
            ],
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

    it('should get project id and permissions', async () => {
        await fixture.whenStable();
        expect(component.projectId).toEqual(1);
        expect(component.hasTagImmutablePermission).toBeTruthy();
        expect(component.hasTagRetentionPermission).toBeTruthy();
    });
});
