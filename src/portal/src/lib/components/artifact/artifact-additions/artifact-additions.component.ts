import { Component, OnInit, Input } from '@angular/core';
import { AdditionLinks } from "../../../../../ng-swagger-gen/models/addition-links";

@Component({
  selector: 'artifact-additions',
  templateUrl: './artifact-additions.component.html',
  styleUrls: ['./artifact-additions.component.scss']
})
export class ArtifactAdditionsComponent implements OnInit {
  @Input() additionLinks: AdditionLinks;
  constructor() { }

  ngOnInit() {
  }
}
