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
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { Project } from '../../../project';
import { ArtifactListPageService } from './artifact-list-page.service';

@Component({
    selector: 'artifact-list-page',
    templateUrl: 'artifact-list-page.component.html',
    styleUrls: ['./artifact-list-page.component.scss'],
})
export class ArtifactListPageComponent implements OnInit {
    projectId: string;
    projectName: string;
    repoName: string;
    referArtifactNameArray: string[] = [];
    depth: string;
    artifactDigest: string;
    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private artifactListPageService: ArtifactListPageService
    ) {
        this.route.params.subscribe(params => {
            this.depth = this.route.snapshot.params['depth'];
            if (this.depth) {
                const arr: string[] = this.depth.split('-');
                this.referArtifactNameArray = arr.slice(0, arr.length - 1);
                this.artifactDigest = this.depth.split('-')[arr.length - 1];
            } else {
                this.referArtifactNameArray = [];
                this.artifactDigest = null;
            }
        });
    }

    ngOnInit() {
        this.projectId = this.route.snapshot.parent.params['id'];
        let resolverData = this.route.snapshot.parent.data;
        if (resolverData) {
            this.projectName = (<Project>resolverData['projectResolver']).name;
        }
        this.repoName = this.route.snapshot.params['repo'];
        this.artifactListPageService.init(+this.projectId);
    }

    watchGoBackEvt(projectId: string | number): void {
        this.router.navigate(['harbor', 'projects', projectId, 'repositories']);
    }
    goProBack(): void {
        this.router.navigate(['harbor', 'projects']);
    }
    backInitRepo() {
        this.router.navigate([
            'harbor',
            'projects',
            this.projectId,
            'repositories',
            this.repoName,
        ]);
    }
    jumpDigest(index: number) {
        const arr: string[] = this.referArtifactNameArray.slice(0, index + 1);
        if (arr && arr.length) {
            this.router.navigate([
                'harbor',
                'projects',
                this.projectId,
                'repositories',
                this.repoName,
                'artifacts-tab',
                'depth',
                arr.join('-'),
            ]);
        } else {
            this.router.navigate([
                'harbor',
                'projects',
                this.projectId,
                'repositories',
                this.repoName,
            ]);
        }
    }
}
