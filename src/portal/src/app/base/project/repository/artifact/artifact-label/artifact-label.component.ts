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
import { Component, Input, OnInit } from '@angular/core';
import { Artifact } from 'ng-swagger-gen/models/artifact';

@Component({
    selector: 'artifact-label',
    templateUrl: './artifact-label.component.html',
    styleUrls: ['./artifact-label.component.scss'],
})
export class ArtifactLabelComponent implements OnInit {
    @Input() artifactDetails: Artifact;
    artifactExtraAttrs: { [key: string]: any } = {};
    type: string;

    constructor() {}

    ngOnInit(): void {
        if (this.artifactDetails.extra_attrs && this.artifactDetails.type) {
            this.artifactExtraAttrs = this.artifactDetails.extra_attrs;
            this.type = this.artifactDetails.type;
        }
    }

    capitalizeFirstLetter(text: string): string {
        if (!text) return text;
        return text.charAt(0).toUpperCase() + text.slice(1);
    }
}
