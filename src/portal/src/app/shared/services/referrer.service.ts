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

import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { CURRENT_BASE_HREF } from '../units/utils';

/**
 * OCI Descriptor from the referrers response
 */
export interface OCIDescriptor {
    mediaType?: string;
    size?: number;
    digest?: string;
    annotations?: { [key: string]: string };
    artifactType?: string;
}

/**
 * OCI Index returned by the referrers API
 */
export interface OCIIndex {
    schemaVersion?: number;
    mediaType?: string;
    manifests?: OCIDescriptor[];
}

// Known signature artifact types
export const SIGNATURE_ARTIFACT_TYPES = [
    'application/vnd.dev.cosign.artifact.sig.v1+json',
    'application/vnd.dev.sigstore.bundle.v0.3+json',
    'application/vnd.cncf.notary.signature',
];

// Known SBOM artifact types
export const SBOM_ARTIFACT_TYPES = [
    'application/vnd.goharbor.harbor.sbom.v1',
];

@Injectable({
    providedIn: 'root',
})
export class ReferrerService {
    constructor(private http: HttpClient) {}

    /**
     * List referrers for an artifact using the OCI referrers API.
     * Calls: GET /api/v2.0/projects/{project}/repositories/{repo}/artifacts/{digest}/referrers
     */
    listReferrers(
        projectName: string,
        repositoryName: string,
        digest: string,
        artifactType?: string
    ): Observable<OCIIndex> {
        const url = `${CURRENT_BASE_HREF}/projects/${projectName}/repositories/${encodeURIComponent(repositoryName)}/artifacts/${digest}/referrers`;
        let params = new HttpParams();
        if (artifactType) {
            params = params.set('artifactType', artifactType);
        }
        return this.http.get<OCIIndex>(url, { params });
    }

    /**
     * Check if an OCI Index contains any signature referrers.
     */
    static hasSignatures(index: OCIIndex): boolean {
        if (!index?.manifests?.length) {
            return false;
        }
        return index.manifests.some(desc =>
            SIGNATURE_ARTIFACT_TYPES.includes(desc.artifactType)
        );
    }

    /**
     * Check if an OCI Index contains any SBOM referrers.
     */
    static hasSbom(index: OCIIndex): boolean {
        if (!index?.manifests?.length) {
            return false;
        }
        return index.manifests.some(desc =>
            SBOM_ARTIFACT_TYPES.includes(desc.artifactType)
        );
    }

    /**
     * Get the first SBOM digest from the referrers index.
     */
    static getSbomDigest(index: OCIIndex): string | undefined {
        if (!index?.manifests?.length) {
            return undefined;
        }
        const sbom = index.manifests.find(desc =>
            SBOM_ARTIFACT_TYPES.includes(desc.artifactType)
        );
        return sbom?.digest;
    }
}
