import { Component, Output, EventEmitter, OnInit } from "@angular/core";
import { Artifact } from "../../../../../ng-swagger-gen/models/artifact";
import { ArtifactService } from "../../../../../ng-swagger-gen/services/artifact.service";
import { ErrorHandler } from "../../../../lib/utils/error-handler";
import { Label } from "../../../../../ng-swagger-gen/models/label";
import { ProjectService } from "../../../../lib/services";
import { ActivatedRoute, Router } from "@angular/router";
import { AppConfigService } from "../../../app-config.service";
import { Project } from "../../project";
import { finalize } from "rxjs/operators";

@Component({
  selector: "artifact-summary",
  templateUrl: "./artifact-summary.component.html",
  styleUrls: ["./artifact-summary.component.scss"],

  providers: []
})
export class ArtifactSummaryComponent implements OnInit {
  tagId: string;
  artifactDigest: string;
  repositoryName: string;
  projectId: string | number;
  referArtifactNameArray: string[] = [];


  labels: Label;
  artifact: Artifact;
  @Output()
  backEvt: EventEmitter<any> = new EventEmitter<any>();
  projectName: string;
  loading: boolean = false;

  constructor(
    private projectService: ProjectService,
    private artifactService: ArtifactService,
    private errorHandler: ErrorHandler,
    private route: ActivatedRoute,
    private appConfigService: AppConfigService,
    private router: Router
  ) {
  }

  get withAdmiral(): boolean {
    return this.appConfigService.getConfig().with_admiral;
  }

  goBack(): void {
    this.router.navigate(["harbor", "projects", this.projectId, "repositories", this.repositoryName]);
  }

  goBackRep(): void {
    this.router.navigate(["harbor", "projects", this.projectId, "repositories"]);
  }

  goBackPro(): void {
    this.router.navigate(["harbor", "projects"]);
  }
  jumpDigest(index: number) {
    const arr: string[] = this.referArtifactNameArray.slice(0, index + 1 );
    if ( arr && arr.length) {
      this.router.navigate(["harbor", "projects", this.projectId, "repositories", this.repositoryName, "depth", arr.join('-')]);
    } else {
      this.router.navigate(["harbor", "projects", this.projectId, "repositories", this.repositoryName]);
    }
  }


  ngOnInit(): void {
    let depth = this.route.snapshot.params['depth'];
    if (depth) {
      this.referArtifactNameArray = depth.split('-');
    }
    this.repositoryName = this.route.snapshot.params["repo"];
    this.artifactDigest = this.route.snapshot.params["digest"];
    this.projectId = this.route.snapshot.params["id"];
    if (this.repositoryName && this.artifactDigest) {
      const resolverData = this.route.snapshot.data;
      if (resolverData) {
        const pro: Project = <Project>resolverData['projectResolver'];
        this.projectName = pro.name;
        if (this.projectName) {
          this.getArtifactDetails();
        }
      }
    }
  }

  getArtifactDetails(): void {
    this.loading = true;
    this.artifactService.getArtifact({
      repositoryName: this.repositoryName,
      reference: this.artifactDigest,
      projectName: this.projectName,
    }).pipe(finalize(() => this.loading = false))
      .subscribe(response => {
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
