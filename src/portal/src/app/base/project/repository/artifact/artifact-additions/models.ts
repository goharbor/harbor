export class ArtifactBuildHistory {
    created: Date;
    created_by: string;
}
export interface ArtifactDependency {
    name: string;
    version: string;
    repository: string;
}
export interface Addition {
    type: string;
    data?: object;
}

export enum ADDITIONS {
    VULNERABILITIES = 'vulnerabilities',
    BUILD_HISTORY = 'build_history',
    SUMMARY = 'readme.md',
    VALUES = 'values.yaml',
    DEPENDENCIES = 'dependencies',
    SBOMS = 'sboms',
}
