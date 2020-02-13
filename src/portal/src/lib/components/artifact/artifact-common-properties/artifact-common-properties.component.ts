import { Component, OnInit, Input, OnChanges, SimpleChanges } from '@angular/core';
import { Artifact } from "../../../../../ng-swagger-gen/models/artifact";

@Component({
  selector: 'artifact-common-properties',
  templateUrl: './artifact-common-properties.component.html',
  styleUrls: ['./artifact-common-properties.component.scss']
})
export class ArtifactCommonPropertiesComponent implements OnInit, OnChanges {
  @Input() artifactDetails: Artifact;
  commonProperties: { [key: string]: any } = {};

  constructor() {
  }

  ngOnInit() {
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes && changes["artifactDetails"]) {
      if (this.artifactDetails) {
        Object.assign(this.commonProperties, this.artifactDetails.extra_attrs, this.artifactDetails.annotations);
        for (let name in this.commonProperties) {
          if (this.commonProperties.hasOwnProperty(name)) {
            this.commonProperties[name] = JSON.stringify(this.commonProperties[name]);
          }
        }
      }
    }
  }
}
