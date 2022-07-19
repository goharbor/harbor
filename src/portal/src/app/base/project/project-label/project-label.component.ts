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
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { SessionService } from '../../../shared/services/session.service';
import { SessionUser } from '../../../shared/entities/session-user';
import { forkJoin } from 'rxjs';
import {
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../shared/services';
import { ErrorHandler } from '../../../shared/units/error-handler';

@Component({
    selector: 'app-project-config',
    templateUrl: './project-label.component.html',
    styleUrls: ['./project-label.component.scss'],
})
export class ProjectLabelComponent implements OnInit {
    projectId: number;
    projectName: string;
    currentUser: SessionUser;
    hasSignedIn: boolean;
    hasProjectAdminRole: boolean;
    hasCreateLabelPermission: boolean;
    hasUpdateLabelPermission: boolean;
    hasDeleteLabelPermission: boolean;
    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private userPermissionService: UserPermissionService,
        private errorHandler: ErrorHandler,
        private session: SessionService
    ) {}

    ngOnInit() {
        this.projectId = +this.route.snapshot.parent.parent.params['id'];
        this.currentUser = this.session.getCurrentUser();
        this.hasSignedIn = this.session.getCurrentUser() !== null;
        this.getLabelPermissionRule(this.projectId);
    }

    getLabelPermissionRule(projectId: number): void {
        const hasCreateLabelPermission =
            this.userPermissionService.getPermission(
                projectId,
                USERSTATICPERMISSION.LABEL.KEY,
                USERSTATICPERMISSION.LABEL.VALUE.CREATE
            );
        const hasUpdateLabelPermission =
            this.userPermissionService.getPermission(
                projectId,
                USERSTATICPERMISSION.LABEL.KEY,
                USERSTATICPERMISSION.LABEL.VALUE.UPDATE
            );
        const hasDeleteLabelPermission =
            this.userPermissionService.getPermission(
                projectId,
                USERSTATICPERMISSION.LABEL.KEY,
                USERSTATICPERMISSION.LABEL.VALUE.DELETE
            );
        forkJoin(
            hasCreateLabelPermission,
            hasUpdateLabelPermission,
            hasDeleteLabelPermission
        ).subscribe(
            permissions => {
                this.hasCreateLabelPermission = permissions[0] as boolean;
                this.hasUpdateLabelPermission = permissions[1] as boolean;
                this.hasDeleteLabelPermission = permissions[2] as boolean;
            },
            error => this.errorHandler.error(error)
        );
    }
}
