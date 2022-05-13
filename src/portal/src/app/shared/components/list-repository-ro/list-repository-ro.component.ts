// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import { ChangeDetectionStrategy, Component, Input } from '@angular/core';
import { Router } from '@angular/router';
import { Repository } from '../../../../../ng-swagger-gen/models/repository';
import { SearchTriggerService } from '../global-search/search-trigger.service';
import { SessionService } from '../../services/session.service';
import { UN_LOGGED_PARAM, YES } from '../../../account/sign-in/sign-in.service';

@Component({
    selector: 'list-repository-ro',
    templateUrl: 'list-repository-ro.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ListRepositoryROComponent {
    @Input() repositories: Repository[];

    constructor(
        private router: Router,
        private searchTrigger: SearchTriggerService,
        private sessionService: SessionService
    ) {}

    getLink(projectId: number, repoName: string) {
        let projectName = repoName.split('/')[0];
        let repositorieName = projectName
            ? repoName.substr(projectName.length + 1)
            : repoName;
        return ['/harbor/projects', projectId, 'repositories', repositorieName];
    }
    getQueryParams() {
        if (this.sessionService.getCurrentUser()) {
            return null;
        }
        return { [UN_LOGGED_PARAM]: YES };
    }
}
