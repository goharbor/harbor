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
import {
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../shared/services';
import { forkJoin, Observable } from 'rxjs';
import { ActivatedRoute } from '@angular/router';

@Component({
    selector: 'app-tag-feature-integration',
    templateUrl: './tag-feature-integration.component.html',
    styleUrls: ['./tag-feature-integration.component.scss'],
})
export class TagFeatureIntegrationComponent implements OnInit {
    projectId: number;
    hasTagRetentionPermission: boolean;
    hasTagImmutablePermission: boolean;
    constructor(
        private userPermissionService: UserPermissionService,
        private route: ActivatedRoute
    ) {}
    ngOnInit() {
        this.projectId = this.route.snapshot.parent.parent.params['id'];
        const permissionsList: Array<Observable<boolean>> = [];
        permissionsList.push(
            this.userPermissionService.getPermission(
                this.projectId,
                USERSTATICPERMISSION.TAG_RETENTION.KEY,
                USERSTATICPERMISSION.TAG_RETENTION.VALUE.READ
            )
        );
        permissionsList.push(
            this.userPermissionService.getPermission(
                this.projectId,
                USERSTATICPERMISSION.IMMUTABLE_TAG.KEY,
                USERSTATICPERMISSION.IMMUTABLE_TAG.VALUE.LIST
            )
        );
        forkJoin(permissionsList).subscribe(Rules => {
            [this.hasTagRetentionPermission, this.hasTagImmutablePermission] =
                Rules;
        });
    }
}
