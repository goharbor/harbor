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
import { Router, ActivatedRoute } from '@angular/router';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { ProjectLabelComponent } from './project-label.component';
import { SessionService } from '../../../shared/services/session.service';
import { UserPermissionService } from '../../../shared/services';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('ProjectLabelComponent', () => {
    let component: ProjectLabelComponent;
    let fixture: ComponentFixture<ProjectLabelComponent>;
    let fakeRouter = null;
    const fakeUserPermissionService = {
        getPermission() {
            return of(true);
        },
    };
    const fakeSessionService = {
        getCurrentUser: function () {
            return { has_admin_role: true };
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ProjectLabelComponent],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            providers: [
                { provide: Router, useValue: fakeRouter },
                {
                    provide: ActivatedRoute,
                    useValue: {
                        snapshot: {
                            parent: {
                                parent: {
                                    params: {
                                        id: 1,
                                    },
                                },
                            },
                        },
                    },
                },
                {
                    provide: UserPermissionService,
                    useValue: fakeUserPermissionService,
                },
                { provide: SessionService, useValue: fakeSessionService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ProjectLabelComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
