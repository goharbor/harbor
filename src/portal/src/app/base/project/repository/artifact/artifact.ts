import { Accessory } from 'ng-swagger-gen/models/accessory';
import { Artifact } from '../../../../../../ng-swagger-gen/models/artifact';
import { Platform } from '../../../../../../ng-swagger-gen/models/platform';
import { Label } from '../../../../../../ng-swagger-gen/models/label';

export interface ArtifactFront extends Artifact {
    platform?: Platform;
    showImage?: string;
    pullCommand?: string;
    annotationsArray?: Array<{ [key: string]: any }>;
    tagNumber?: number;
    coSigned?: string;
    accessoryNumber?: number;
}

export interface AccessoryFront extends Accessory {
    coSigned?: string;
    accessoryNumber?: number;
    accessories?: any;
}

export const multipleFilter: Array<{
    filterBy: string;
    filterByShowText: string;
    listItem: any[];
}> = [
    {
        filterBy: 'type',
        filterByShowText: 'Type',
        listItem: [
            {
                filterText: 'IMAGE',
                showItem: 'ARTIFACT.IMAGE',
            },
            {
                filterText: 'CHART',
                showItem: 'ARTIFACT.CHART',
            },
            {
                filterText: 'CNAB',
                showItem: 'ARTIFACT.CNAB',
            },
            {
                filterText: 'WASM',
                showItem: 'ARTIFACT.WASM',
            },
        ],
    },
    {
        filterBy: 'tags',
        filterByShowText: 'Tags',
        listItem: [
            {
                filterText: '*',
                showItem: 'ARTIFACT.TAGGED',
            },
            {
                filterText: 'nil',
                showItem: 'ARTIFACT.UNTAGGED',
            },
            {
                filterText: '',
                showItem: 'ARTIFACT.ALL',
            },
        ],
    },
    {
        filterBy: 'labels',
        filterByShowText: 'Label',
        listItem: [],
    },
];

export enum AccessoryType {
    COSIGN = 'signature.cosign',
    NYDUS = 'accelerator.nydus',
}

export enum ArtifactType {
    IMAGE = 'IMAGE',
    CHART = 'CHART',
    CNAB = 'CNAB',
    OPENPOLICYAGENT = 'OPENPOLICYAGENT',
}

export const artifactDefault = 'images/artifact-default.svg';

export enum AccessoryQueryParams {
    ACCESSORY_TYPE = 'accessoryType',
}

export function hasPullCommand(artifact: Artifact): boolean {
    return (
        artifact.type === ArtifactType.IMAGE ||
        artifact.type === ArtifactType.CNAB ||
        artifact.type === ArtifactType.CHART
    );
}

export function getPullCommandByDigest(
    artifactType: string,
    url: string,
    digest: string,
    client: Clients
): string {
    if (artifactType && url && digest) {
        if (artifactType === ArtifactType.IMAGE) {
            if (client === Clients.DOCKER) {
                return `${Clients.DOCKER} pull ${url}@${digest}`;
            }
            if (client === Clients.PODMAN) {
                return `${Clients.PODMAN} pull ${url}@${digest}`;
            }
        }
        if (artifactType === ArtifactType.CNAB) {
            return `${Clients.CNAB} pull ${url}@${digest}`;
        }
    }
    return null;
}

export function getPullCommandByTag(
    artifactType: string,
    url: string,
    tag: string,
    client: Clients
): string {
    if (artifactType && url && tag) {
        if (artifactType === ArtifactType.IMAGE) {
            if (client === Clients.DOCKER) {
                return `${Clients.DOCKER} pull ${url}:${tag}`;
            }
            if (client === Clients.PODMAN) {
                return `${Clients.PODMAN} pull ${url}:${tag}`;
            }
        }
        if (artifactType === ArtifactType.CNAB) {
            return `cnab-to-oci pull ${url}:${tag}`;
        }
        if (artifactType === ArtifactType.CHART) {
            return `helm pull oci://${url} --version ${tag}`;
        }
    }
    return null;
}

export interface ArtifactFilterEvent {
    type?: string;
    stringValue?: string;
    isLabel?: boolean;
    isInputTag?: boolean;
    label?: Label;
}

export enum Clients {
    DOCKER = 'docker',
    PODMAN = 'podman',
    CHART = 'helm',
    CNAB = 'cnab-to-oci',
}

export enum ClientNames {
    DOCKER = 'Docker',
    PODMAN = 'Podman',
    CHART = 'Helm',
    CNAB = 'CNAB',
}
