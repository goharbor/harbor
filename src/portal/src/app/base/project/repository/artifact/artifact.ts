import { Accessory } from "ng-swagger-gen/models/accessory";
import { Artifact } from "../../../../../../ng-swagger-gen/models/artifact";
import { Platform } from "../../../../../../ng-swagger-gen/models/platform";

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
  pullCommand?: string;
  tagNumber?: number;
  scan_overview?: any;
}

export const mutipleFilter = [
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
      }
    ]
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
      }
    ]
  },
  {
    filterBy: 'labels',
    filterByShowText: 'Label',
    listItem: []
  },
];

export enum AccessoryType {
  COSIGN = 'signature.cosign'
}

export const artifactImages = [
  'IMAGE', 'CHART', 'CNAB', 'OPENPOLICYAGENT'
];
export const artifactPullCommands = [
  {
    type: artifactImages[0],
    pullCommand: 'docker pull'
  },
  {
    type: AccessoryType.COSIGN,
    pullCommand: 'docker pull'
  },
  {
    type: artifactImages[1],
    pullCommand: 'helm chart pull'
  },
  {
    type: artifactImages[2],
    pullCommand: 'cnab-to-oci pull'
  }
];
export const artifactDefault = "images/artifact-default.svg";



