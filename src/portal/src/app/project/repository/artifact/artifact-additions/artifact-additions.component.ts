import { Component, OnInit, Input } from '@angular/core';
import { ADDITIONS } from "./models";
import { AdditionLinks } from "../../../../../../ng-swagger-gen/models/addition-links";
import { AdditionLink } from "../../../../../../ng-swagger-gen/models/addition-link";

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
  getVulnerability(): AdditionLink {
    if (this.additionLinks && this.additionLinks[ADDITIONS.VULNERABILITIES]) {
      return this.additionLinks[ADDITIONS.VULNERABILITIES];
    }
    return null;
  }
  getBuildHistory(): AdditionLink {
    if (this.additionLinks && this.additionLinks[ADDITIONS.BUILD_HISTORY]) {
      return this.additionLinks[ADDITIONS.BUILD_HISTORY];
    }
    return null;
  }
  getSummary(): AdditionLink {
    if (this.additionLinks && this.additionLinks[ADDITIONS.SUMMARY]) {
      return this.additionLinks[ADDITIONS.SUMMARY];
    }
    return null;
  }
  getDependencies(): AdditionLink {
    if (this.additionLinks && this.additionLinks[ADDITIONS.DEPENDENCIES]) {
      return this.additionLinks[ADDITIONS.DEPENDENCIES];
    }
    return null;
  }
  getValues(): AdditionLink {
    if (this.additionLinks && this.additionLinks[ADDITIONS.VALUES]) {
      return this.additionLinks[ADDITIONS.VALUES];
    }
    return null;
  }
}
