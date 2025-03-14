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
import { ArtifactTagComponent } from './artifact-tag.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { ArtifactService } from '../../../../../../../ng-swagger-gen/services/artifact.service';
import { OperationService } from '../../../../../shared/components/operation/operation.service';
import {
    SystemInfoService,
    UserPermissionDefaultService,
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../../../shared/services';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../../../../shared/shared.module';

describe('ArtifactTagComponent', () => {
    let component: ArtifactTagComponent;
    let fixture: ComponentFixture<ArtifactTagComponent>;
    const mockErrorHandler = {
        error: () => {},
    };
    const mockArtifactService = {
        createTag: () => of([]),
        deleteTag: () => of(null),
        listTagsResponse: () => of([]).pipe(delay(0)),
    };
    const mockSystemInfoService = {
        getSystemInfo: () => of(false),
    };

    let userPermissionService;
    const permissions = [
        {
            resource: USERSTATICPERMISSION.REPOSITORY_TAG.KEY,
            action: USERSTATICPERMISSION.REPOSITORY_TAG.VALUE.DELETE,
        },
        {
            resource: USERSTATICPERMISSION.REPOSITORY_TAG.KEY,
            action: USERSTATICPERMISSION.REPOSITORY_TAG.VALUE.CREATE,
        },
    ];
    let mockHasDeleteImagePermission: boolean = true;
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            declarations: [ArtifactTagComponent],
            providers: [
                { provide: ErrorHandler, useValue: mockErrorHandler },
                { provide: ArtifactService, useValue: mockArtifactService },
                { provide: SystemInfoService, useValue: mockSystemInfoService },
                {
                    provide: UserPermissionService,
                    useClass: UserPermissionDefaultService,
                },
                { provide: OperationService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactTagComponent);
        component = fixture.componentInstance;
        userPermissionService = fixture.debugElement.injector.get(
            UserPermissionService
        );
        spyOn(userPermissionService, 'hasProjectPermissions')
            .withArgs(component.projectId, permissions)
            .and.returnValue(of([mockHasDeleteImagePermission]));
        component.artifactDetails = { id: 1 };
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
