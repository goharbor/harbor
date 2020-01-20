import { Component, OnInit, Input } from '@angular/core';
import { Artifact } from '../artifact';

@Component({
  selector: 'artifact-additional-link',
  templateUrl: './artifact-additional-link.component.html',
  styleUrls: ['./artifact-additional-link.component.scss']
})
export class ArtifactAdditionalLinkComponent implements OnInit {
  @Input() artifactDetails: Artifact;

  constructor() { }

  ngOnInit() {
  }

}
