/*
 {
    "tag": "latest",
    "manifest": {
      "schemaVersion": 1,
      "name": "library/photon",
      "tag": "latest",
      "architecture": "amd64",
      "history": []
    },

*/
export class Tag {
  tag: string;
  manifest: {
    schemaVersion: number;
    name: string;
    tag: string;
    architecture: string;
    history: [
      {
        v1Compatibility: string;
      }
    ];
  };
  verified: boolean;
}