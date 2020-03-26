import { Artifact } from "../../../../../ng-swagger-gen/models/artifact";

export interface ArtifactFront extends Artifact {
    annotationsArray?: string[];
    showImage?: string;
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
