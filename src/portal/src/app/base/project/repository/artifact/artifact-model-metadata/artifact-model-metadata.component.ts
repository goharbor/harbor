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
import { Component, Input, OnChanges } from '@angular/core';
import { Artifact } from 'ng-swagger-gen/models/artifact';

export const CNAI_TYPE = 'CNAI';
// extra_attrs keys of a CNAI artifact (CNCF model-spec config) rendered by this
// panel instead of as raw JSON in the overview
export const CNAI_METADATA_KEYS = ['descriptor', 'config', 'modelfs'];

interface Row {
    key: string;
    value: string;
    href?: string;
}

const SAFE_URL = /^https?:\/\//i;

@Component({
    selector: 'artifact-model-metadata',
    templateUrl: './artifact-model-metadata.component.html',
    styleUrls: ['./artifact-model-metadata.component.scss'],
    standalone: false,
})
export class ArtifactModelMetadataComponent implements OnChanges {
    @Input() artifactDetails: Artifact;
    rows: Row[] = [];
    licenses: string[] = [];

    ngOnChanges(): void {
        this.rows = [];
        this.licenses = [];
        if (this.artifactDetails?.type !== CNAI_TYPE) {
            return;
        }
        const attrs: any = this.artifactDetails.extra_attrs || {};
        const d = attrs.descriptor || {};
        const c = attrs.config || {};
        const fs = attrs.modelfs || {};
        this.licenses = Array.isArray(d.licenses) ? d.licenses : [];
        const source = SAFE_URL.test(d.sourceURL || '') ? d.sourceURL : '';
        this.add('ARTIFACT.MODEL_NAME', d.title || d.name);
        this.add('ARTIFACT.MODEL_FAMILY', d.family);
        this.add('ARTIFACT.MODEL_VERSION', d.version);
        this.add('ARTIFACT.MODEL_SOURCE', d.sourceURL, source);
        this.add('ARTIFACT.MODEL_REVISION', d.revision);
        this.add('ARTIFACT.MODEL_ARCHITECTURE', c.architecture);
        this.add('ARTIFACT.MODEL_FORMAT', c.format);
        this.add('ARTIFACT.MODEL_PARAMS', c.paramSize);
        this.add('ARTIFACT.MODEL_PRECISION', c.precision);
        this.add('ARTIFACT.MODEL_QUANTIZATION', c.quantization);
        this.add(
            'ARTIFACT.MODEL_AUTHORS',
            Array.isArray(d.authors) ? d.authors.join(', ') : d.authors
        );
        this.add('ARTIFACT.MODEL_VENDOR', d.vendor);
        this.add('ARTIFACT.MODEL_CREATED', d.createdAt);
        this.add(
            'ARTIFACT.MODEL_LAYERS',
            Array.isArray(fs.diffIds) && fs.diffIds.length
                ? String(fs.diffIds.length)
                : ''
        );
        if (SAFE_URL.test(d.docURL || '')) {
            this.add('ARTIFACT.MODEL_DOCS', d.docURL, d.docURL);
        }
    }

    private add(key: string, value: any, href?: string): void {
        if (value === undefined || value === null || value === '') {
            return;
        }
        this.rows.push({ key, value: String(value), href: href || undefined });
    }
}
