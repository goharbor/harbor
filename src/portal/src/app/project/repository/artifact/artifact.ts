import { Artifact } from "../../../../../ng-swagger-gen/models/artifact";
import { Platform } from "../../../../../ng-swagger-gen/models/platform";

export interface ArtifactFront extends Artifact {
    platform?: Platform;
    showImage?: string;
    pullCommand?: string;
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
  export const artifactImages = [
      'IMAGE', 'CHART', 'CNAB', 'OPENPOLICYAGENT'
  ];
  export const artifactPullCommands = [
    {
      type: artifactImages[0],
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
