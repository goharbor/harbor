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
import { Component, Input, OnChanges, SimpleChanges } from '@angular/core';
import { DatePipe } from '@angular/common';
import { Artifact } from '../../../../../../../ng-swagger-gen/models/artifact';

enum Types {
    CREATED = 'created',
    TYPE = 'type',
    MEDIA_TYPE = 'media_type',
    MANIFEST_MEDIA_TYPE = 'manifest_media_type',
    DIGEST = 'digest',
    SIZE = 'size',
    PUSH_TIME = 'push_time',
    PULL_TIME = 'pull_time',
}

@Component({
    selector: 'artifact-common-properties',
    templateUrl: './artifact-common-properties.component.html',
    styleUrls: ['./artifact-common-properties.component.scss'],
})
export class ArtifactCommonPropertiesComponent implements OnChanges {
    @Input() artifactDetails: Artifact;
    commonProperties: { [key: string]: any } = {};

    constructor() {}

    ngOnChanges(changes: SimpleChanges) {
        if (changes && changes['artifactDetails']) {
            if (this.artifactDetails) {
                Object.assign(
                    this.commonProperties,
                    this.artifactDetails.extra_attrs,
                    this.artifactDetails.annotations
                );
                for (let name in this.commonProperties) {
                    if (this.commonProperties.hasOwnProperty(name)) {
                        if (typeof this.commonProperties[name] === 'object') {
                            if (this.commonProperties[name] === null) {
                                this.commonProperties[name] = '';
                            } else {
                                this.commonProperties[name] = JSON.stringify(
                                    this.commonProperties[name]
                                );
                            }
                        }
                        if (name === Types.CREATED) {
                            this.commonProperties[name] = new DatePipe(
                                'en-us'
                            ).transform(this.commonProperties[name], 'short');
                        }
                    }
                }
            }
        }
    }

    hasCommonProperties(): boolean {
        return JSON.stringify(this.commonProperties) !== '{}';
    }
}
