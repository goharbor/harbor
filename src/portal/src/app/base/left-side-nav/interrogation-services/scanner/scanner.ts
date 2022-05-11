import { ScannerRegistration } from '../../../../../../ng-swagger-gen/models/scanner-registration';
import { ScannerAdapterMetadata } from '../../../../../../ng-swagger-gen/models/scanner-adapter-metadata';

export interface Scanner extends ScannerRegistration {
    metadata?: ScannerAdapterMetadata;
    loadingMetadata?: boolean;
}

export const SCANNERS_DOC: string =
    'https://goharbor.io/blog/harbor-1.10-release/#vulnerability-scanning-with-pluggable-scanners';
