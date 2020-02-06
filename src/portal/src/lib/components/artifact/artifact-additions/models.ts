import { HelmChartMaintainer } from "../../../../app/project/helm-chart/helm-chart.interface.service";

export class ArtifactBuildHistory {
    createdTime: Date;
    createdBy: string;
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

export const ADDITIONS = {
    BUILD_HISTORY: {
        i18nKey: 'ARTIFACT.BUILD_HISTORY',
        type: 'json'
    },
    SUMMARY: {
        i18nKey: 'ARTIFACT.SUMMARY',
        type: 'markdown'
    },
    DEPENDENCIES: {
        i18nKey: 'ARTIFACT.DEPENDENCIES',
        type: 'yaml'
    },
    VALUES: {
        i18nKey: 'ARTIFACT.VALUES',
        type: 'yaml'
    }
};

