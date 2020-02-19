import { Component, Input, Output, EventEmitter, OnInit } from "@angular/core";
import { ProjectService } from "../../services";
import { ErrorHandler } from "../../utils/error-handler";
import { Label } from "../../services/interface";
import { Artifact } from "../../../../ng-swagger-gen/models/artifact";
import { ArtifactService } from "../../../../ng-swagger-gen/services/artifact.service";

@Component({
  selector: "artifact-summary",
  templateUrl: "./artifact-summary.component.html",
  styleUrls: ["./artifact-summary.component.scss"],

  providers: []
})
export class ArtifactSummaryComponent implements OnInit {
  labels: Label;
  @Input()
  artifactDigest: string;
  @Input()
  repositoryName: string;
  @Input()
  withAdmiral: boolean;
  artifact: Artifact;
  @Output()
  backEvt: EventEmitter<any> = new EventEmitter<any>();
  @Input() projectId: number;
  projectName: string;


  constructor(
    private projectService: ProjectService,
    private artifactService: ArtifactService,
    private errorHandler: ErrorHandler,
  ) {
  }

  ngOnInit(): void {
    if (this.repositoryName && this.artifactDigest) {
      this.projectService.getProject(this.projectId).subscribe(project => {
        this.projectName = project.name;
        this.getArtifactDetails();
      });
    }
  }

  getArtifactDetails(): void {
    this.artifactService.getArtifact({
      repositoryName: this.repositoryName,
      reference: this.artifactDigest,
      projectName: this.projectName,
      withLabel: true,
      withScanOverview: true,
      withSignature: true,
      withImmutableStatus: true
    }).subscribe(response => {
      this.artifact = response;
    }, error => {
      this.errorHandler.error(error);
    });
  }

  onBack(): void {
    this.backEvt.emit(this.repositoryName);
  }

  refreshArtifact() {
    this.getArtifactDetails();
  }
}
