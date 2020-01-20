import { Component, OnInit, Input, OnChanges, SimpleChanges } from '@angular/core';
import { Artifact } from '../artifact';

@Component({
  selector: 'artifact-common-properties',
  templateUrl: './artifact-common-properties.component.html',
  styleUrls: ['./artifact-common-properties.component.scss']
})
export class ArtifactCommonPropertiesComponent implements OnInit, OnChanges {
  @Input() artifactDetails: Artifact;
  constructor() { }

  ngOnInit() {
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes && changes["artifactDetails"]) {
        // this.originalConfig = clone(this.currentConfig);
    }
}

}
