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

/* tslint:disable */

import { Scanner } from 'ng-swagger-gen/models';

/**
 * The generate SBOM overview information
 */
export interface SBOMOverview {
    /**
     * id of the native sbom report
     */
    report_id?: string;

    /**
     * The start time of the scan process that generating report
     */
    start_time?: string;

    /**
     * The end time of the scan process that generating report
     */
    end_time?: string;

    /**
     * The status of the generate SBOM process
     */
    scan_status?: string;

    /**
     * The digest of the generated SBOM accessory
     */
    sbom_digest?: string;

    /**
     * The seconds spent for generating the report
     */
    duration?: number;
    scanner?: Scanner;
}
