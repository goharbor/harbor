import { Component, OnInit, Input } from '@angular/core';
import { ExtraAttrs } from "../../../../../ng-swagger-gen/models/extra-attrs";
import {AdditionsService} from "./additions.service";

@Component({
  selector: 'artifact-additions',
  templateUrl: './artifact-additions.component.html',
  styleUrls: ['./artifact-additions.component.scss']
})
export class ArtifactAdditionsComponent implements OnInit {
  @Input() extraAttrs: ExtraAttrs;
  constructor() { }

  ngOnInit() {
  }
}
