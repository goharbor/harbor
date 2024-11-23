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
    signed?: string;
    sbomDigest?: string;
    accessoryNumber?: number;
    accessoryLoading?: boolean;
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
    NOTATION = 'signature.notation',
    NYDUS = 'accelerator.nydus',
    SBOM = 'sbom.harbor',
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
            if (Object.values(Clients).includes(client)) {
                return `${client} pull ${url}@${digest}`;
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
            if (Object.values(Clients).includes(client)) {
                return `${client} pull ${url}:${tag}`;
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
    NERDCTL = 'nerdctl',
    CONTAINERD = 'ctr',
    CRI_O = 'crictl',
    CHART = 'helm',
    CNAB = 'cnab-to-oci',
}

export enum ClientNames {
    DOCKER = 'Docker',
    PODMAN = 'Podman',
    NERDCTL = 'nerdctl',
    CONTAINERD = 'ctr',
    CRI_O = 'crictl',
    CHART = 'Helm',
    CNAB = 'CNAB',
}

export enum ArtifactSbomType {
    SPDX = 'SPDX',
}

export interface ArtifactSbomPackageItem {
    name?: string;
    versionInfo?: string;
    licenseConcluded?: string;
    [key: string]: Object;
}

export interface ArtifactSbomPackage {
    packages: ArtifactSbomPackageItem[];
}

export interface ArtifactSbom {
    sbomType: ArtifactSbomType;
    sbomVersion: string;
    sbomName?: string;
    sbomDataLicense?: string;
    sbomId?: string;
    sbomDocumentNamespace?: string;
    sbomCreated?: string;
    sbomPackage?: ArtifactSbomPackage;
    sbomJsonRaw?: Object;
}

export const ArtifactSbomFieldMapper = {
    sbomVersion: 'spdxVersion',
    sbomName: 'name',
    sbomDataLicense: 'dataLicense',
    sbomId: 'SPDXID',
    sbomDocumentNamespace: 'documentNamespace',
    sbomCreated: 'creationInfo.created',
    sbomPackage: {
        packages: ['name', 'versionInfo', 'licenseConcluded'],
    },
};

/**
 * Identify the sbomJson contains the two main properties 'spdxVersion' and 'SPDXID'.
 * @param sbomJson SBOM JSON report object.
 * @returns true or false
 * Return true when the sbomJson object contains the attribues 'spdxVersion' and 'SPDXID'.
 * else return false.
 */
export function isSpdxSbom(sbomJson?: Object): boolean {
    return Object.keys(sbomJson ?? {}).includes(ArtifactSbomFieldMapper.sbomId);
}

/**
 * Update the value to the data object with the field path.
 * @param fieldPath field class path eg {a: {b:'test'}}. field path for b is 'a.b'
 * @param data The target object to receive the value.
 * @param value The value will be set to the data object.
 */
export function updateObjectWithFieldPath(
    fieldPath: string,
    data: Object,
    value: Object
) {
    if (fieldPath && data) {
        const fields = fieldPath?.split('.');
        let tempData = data;
        fields.forEach((field, index) => {
            const properties = Object.getOwnPropertyNames(tempData);
            if (field !== '__proto__' && field !== 'constructor') {
                if (index === fields.length - 1) {
                    tempData[field] = value;
                } else {
                    if (!properties.includes(field)) {
                        tempData[field] = {};
                    }
                    tempData = tempData[field];
                }
            }
        });
    }
}

/**
 * Get value from data object with field path.
 * @param fieldPath field class path eg {a: {b:'test'}}. field path for b is 'a.b'
 * @param data The data source target object.
 * @returns The value read from data object.
 */
export const getValueFromObjectWithFieldPath = (
    fieldPath: string,
    data: Object
) => {
    let tempObject = data;
    if (fieldPath && data) {
        const fields = fieldPath?.split('.');
        fields.forEach(field => {
            if (tempObject) {
                tempObject = tempObject[field] ?? null;
            }
        });
    }
    return tempObject;
};

/**
 * Get value from source data object with field path.
 * @param fieldPathObject The Object that contains the field paths.
 * If we have an Object - {a: {b: 'test', c: [{ d: 2, e: 'v'}]}}.
 * The field path for b is 'a.b'.
 * The field path for c is {'a.c': ['d', 'e']'}.
 * @param sourceData The data source target object.
 * @returns the value by field class path.
 */
export function readDataFromArtifactSbomJson(
    fieldPathObject: Object,
    sourceData: Object
): Object {
    let result = null;
    if (sourceData) {
        switch (typeof fieldPathObject) {
            case 'string':
                result = getValueFromObjectWithFieldPath(
                    fieldPathObject,
                    sourceData
                );
                break;
            case 'object':
                if (
                    Array.isArray(fieldPathObject) &&
                    Array.isArray(sourceData)
                ) {
                    result = sourceData.map(source => {
                        let arrayItem = {};
                        fieldPathObject.forEach(field => {
                            updateObjectWithFieldPath(
                                field,
                                arrayItem,
                                readDataFromArtifactSbomJson(field, source)
                            );
                        });
                        return arrayItem;
                    });
                } else {
                    const fields = Object.getOwnPropertyNames(fieldPathObject);
                    result = result ? result : {};
                    fields.forEach(field => {
                        if (sourceData[field]) {
                            updateObjectWithFieldPath(
                                field,
                                result,
                                readDataFromArtifactSbomJson(
                                    fieldPathObject[field],
                                    sourceData[field]
                                )
                            );
                        }
                    });
                }
                break;
            default:
                break;
        }
    }
    return result;
}

/**
 * Convert  SBOM Json report to ArtifactSbom
 * @param sbomJson SBOM report in Json format
 * @returns ArtifactSbom || null
 */
export function getArtifactSbom(sbomJson?: Object): ArtifactSbom {
    if (sbomJson) {
        if (isSpdxSbom(sbomJson)) {
            const artifactSbom = <ArtifactSbom>{};
            artifactSbom.sbomJsonRaw = sbomJson;
            artifactSbom.sbomType = ArtifactSbomType.SPDX;
            // only retrieve the fields defined in ArtifactSbomFieldMapper
            const fields = Object.getOwnPropertyNames(ArtifactSbomFieldMapper);
            fields.forEach(field => {
                updateObjectWithFieldPath(
                    field,
                    artifactSbom,
                    readDataFromArtifactSbomJson(
                        ArtifactSbomFieldMapper[field],
                        sbomJson
                    )
                );
            });
            return artifactSbom;
        }
    }
    return null;
}
