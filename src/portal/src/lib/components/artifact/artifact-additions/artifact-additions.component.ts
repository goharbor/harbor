import { Component, OnInit, Input } from '@angular/core';
import { Artifact } from '../artifact';

@Component({
  selector: 'artifact-additions',
  templateUrl: './artifact-additions.component.html',
  styleUrls: ['./artifact-additions.component.scss']
})
export class ArtifactAdditionsComponent implements OnInit {
  @Input() artifactDetails: Artifact;

  constructor() { }

  ngOnInit() {
  }

}
