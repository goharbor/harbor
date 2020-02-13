import { HelmChartMaintainer } from "../../../../app/project/helm-chart/helm-chart.interface.service";

export class ArtifactBuildHistory {
    created: Date;
    created_by: string;
}
export interface ArtifactDependency {
    name: string;
    version: string;
    repository: string;
}
export interface ArtifactSummary {
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


export interface Addition {
    type: string;
    data?: object;
}

export enum ADDITIONS  {
    VULNERABILITIES = 'vulnerabilities',
    BUILD_HISTORY = 'build_history',
    SUMMARY = 'readme',
    VALUES = 'values.yaml',
    DEPENDENCIES = 'dependencies'
}

