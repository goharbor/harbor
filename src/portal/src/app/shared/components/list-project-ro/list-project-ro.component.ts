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
import { Project } from '../../../../../ng-swagger-gen/models/project';

@Component({
    selector: 'list-project-ro',
    templateUrl: 'list-project-ro.component.html',
    changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ListProjectROComponent {
    @Input() projects: Project[];

    constructor() {}

    getLink(proId: number) {
        return `/harbor/projects/${proId}/repositories`;
    }
}
