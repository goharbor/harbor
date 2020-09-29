import {
  Component,
  OnInit,
  Input,
} from "@angular/core";
import { ArtifactDependency } from "../models";
import { AdditionsService } from "../additions.service";
import { AdditionLink } from "../../../../../../../ng-swagger-gen/models/addition-link";
import { ErrorHandler } from "../../../../../../lib/utils/error-handler";
import { pipe } from "rxjs";
import { finalize } from "rxjs/operators";


@Component({
  selector: "hbr-artifact-dependencies",
  templateUrl: "./dependencies.component.html",
  styleUrls: ["./dependencies.component.scss"],
})
export class DependenciesComponent implements OnInit {
  @Input()
  dependenciesLink: AdditionLink;
  dependencyList: ArtifactDependency[] = [];
  loading: boolean = false;
  constructor( private errorHandler: ErrorHandler,
               private additionsService: AdditionsService) {}

  ngOnInit(): void {
    this.getDependencyList();
  }
  getDependencyList() {
    if (this.dependenciesLink
        && !this.dependenciesLink.absolute
        && this.dependenciesLink.href) {
      this.loading = true;
      this.additionsService.getDetailByLink(this.dependenciesLink.href)
        .pipe(finalize(() => this.loading = false))
          .subscribe(
        res => {
          this.dependencyList = res;
        }, error => {
          this.errorHandler.error(error);
        }
      );
    }
  }
}
