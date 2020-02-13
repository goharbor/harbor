import { Component, Input, OnInit } from "@angular/core";
import { ErrorHandler } from "../../../../utils/error-handler";
import { AdditionsService } from "../additions.service";
import { ArtifactBuildHistory } from "../models";
import { AdditionLink } from "../../../../../../ng-swagger-gen/models/addition-link";
import { finalize } from "rxjs/operators";

@Component({
  selector: "hbr-artifact-build-history",
  templateUrl: "./build-history.component.html",
  styleUrls: ["./build-history.component.scss"],
})
export class BuildHistoryComponent implements OnInit {
  @Input()
  buildHistoryLink: AdditionLink;
  historyList: ArtifactBuildHistory[] = [];
  loading: Boolean = false;
  constructor(
    private errorHandler: ErrorHandler,
    private additionsService: AdditionsService
  ) {
  }

  ngOnInit(): void {
    this.getBuildHistory();
  }
  getBuildHistory() {
    if (this.buildHistoryLink
      && !this.buildHistoryLink.absolute
      && this.buildHistoryLink.href) {
      this.loading = true;
      this.additionsService.getDetailByLink(this.buildHistoryLink.href)
        .pipe(finalize(() => this.loading = false))
        .subscribe(
          res => {
            if (res && res.length) {
              res.forEach((ele: any) => {
                const history: ArtifactBuildHistory = new ArtifactBuildHistory();
                history.created = ele.created;
                if (ele.created_by !== undefined) {
                  history.created_by = ele.created_by
                    .replace("/bin/sh -c #(nop)", "")
                    .trimLeft()
                    .replace("/bin/sh -c", "RUN");
                } else {
                  history.created_by = ele.comment;
                }
                this.historyList.push(history);
              });
            }
          }, error => {
            this.errorHandler.error(error);
          }
        );
    }
  }
}
