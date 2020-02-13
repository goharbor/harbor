import { Component, OnInit, Input, SimpleChanges, OnChanges } from '@angular/core';
import { Artifact } from '../artifact';

@Component({
  selector: 'artifact-local-properties',
  templateUrl: './artifact-local-properties.component.html',
  styleUrls: ['./artifact-local-properties.component.scss']
})
export class ArtifactLocalPropertiesComponent implements OnInit, OnChanges {
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
