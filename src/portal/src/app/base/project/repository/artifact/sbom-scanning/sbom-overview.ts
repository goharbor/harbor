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
