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
import { ArtifactListPageComponent } from './artifact-list-page.component';
import { of } from 'rxjs';
import { ActivatedRoute } from '@angular/router';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { ArtifactListPageService } from './artifact-list-page.service';

describe('ArtifactListPageComponent', () => {
    let component: ArtifactListPageComponent;
    let fixture: ComponentFixture<ArtifactListPageComponent>;
    const mockActivatedRoute = {
        RouterparamMap: of({ get: key => 'value' }),
        snapshot: {
            params: {
                id: 1,
            },
            parent: {
                params: { id: 1 },
            },
            data: {
                projectResolver: {
                    has_project_admin_role: true,
                    current_user_role_id: 3,
                },
            },
        },
        data: of({
            projectResolver: {
                ismember: true,
                role_name: 'maintainer',
            },
        }),
        params: {
            subscribe: () => {
                return {
                    unsubscribe() {
                        return null;
                    },
                };
            },
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ArtifactListPageComponent],
            providers: [
                ArtifactListPageService,
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactListPageComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should have two tabs', async () => {
        await fixture.whenStable();
        const tabs = fixture.nativeElement.querySelectorAll('.nav-item');
        expect(tabs.length).toEqual(2);
    });
});
