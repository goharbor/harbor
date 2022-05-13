import { Label } from '../../../../shared/services';

export interface HelmChartSearchResultItem {
    Name: string;
    Score: number;
    Chart: HelmChartVersion;
}
export interface HelmChartItem {
    name: string;
    total_versions: number;
    latest_version: string;
    created: string;
    updated: string;
    icon: string;
    home: string;
    deprecated?: boolean;
    status?: string;
    pulls?: number;
    maintainer?: string;
}

export interface HelmChartVersion {
    name: string;
    home: string;
    sources: string[];
    version: string;
    description: string;
    keywords: string[];
    maintainers: HelmChartMaintainer[];
    engine: string;
    icon: string;
    appVersion: string;
    apiVersion: string;
    urls: string[];
    created: string;
    digest: string;
    labels: Label[];
    deprecated?: boolean;
}

export interface HelmChartDetail {
    metadata: HelmChartMetaData;
    dependencies: HelmChartDependency[];
    values: any;
    files: HelmchartFile;
    security: HelmChartSecurity;
    labels: Label[];
}

export interface HelmChartMetaData {
    name: string;
    home: string;
    sources: string[];
    version: string;
    description: string;
    keywords: string[];
    maintainers: HelmChartMaintainer[];
    engine: string;
    icon: string;
    appVersion: string;
    urls: string[];
    created?: string;
    digest: string;
}

export interface HelmChartMaintainer {
    name: string;
    email: string;
}

export interface HelmChartDependency {
    name: string;
    version: string;
    repository: string;
}

export interface HelmchartFile {
    'README.MD': string;
    'values.yaml': string;
}

export interface HelmChartSecurity {
    signature: HelmChartSignature;
}

export interface HelmChartSignature {
    signed: boolean;
    prov_file: string;
}
