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
import { Project } from '../project';

@Component({
    selector: 'app-project-config',
    templateUrl: './project-config.component.html',
    styleUrls: ['./project-config.component.scss'],
})
export class ProjectConfigComponent implements OnInit {
    projectId: number;
    projectName: string;
    currentUser: SessionUser;
    hasSignedIn: boolean;
    isProxyCacheProject: boolean = false;
    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private session: SessionService
    ) {}

    ngOnInit() {
        this.projectId = +this.route.snapshot.parent.parent.params['id'];
        this.currentUser = this.session.getCurrentUser();
        this.hasSignedIn = this.session.getCurrentUser() !== null;
        let resolverData = this.route.snapshot.parent.parent.data;
        if (resolverData) {
            let pro: Project = <Project>resolverData['projectResolver'];
            this.projectName = pro.name;
            if (pro.registry_id) {
                this.isProxyCacheProject = true;
            }
        }
    }
}
