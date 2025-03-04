import { Component, Input, OnInit } from '@angular/core';
import { Artifact } from 'ng-swagger-gen/models/artifact';

@Component({
    selector: 'artifact-label',
    templateUrl: './artifact-label.component.html',
    styleUrls: ['./artifact-label.component.scss'],
})
export class ArtifactLabelComponent implements OnInit {
    @Input() artifactDetails: Artifact;
    artifactExtraAttrs: { [key: string]: any } = {};
    type: string;

    constructor() {}

    ngOnInit(): void {
        if (this.artifactDetails.extra_attrs && this.artifactDetails.type) {
            this.artifactExtraAttrs = this.artifactDetails.extra_attrs;
            this.type = this.artifactDetails.type;
        }
    }
}
