import {
  Component,
  OnInit,
  Input,
} from "@angular/core";
import { ArtifactDependency } from "../models";
import { ErrorHandler } from "../../../../utils/error-handler";
import { AdditionsService } from "../additions.service";
import { AdditionLink } from "../../../../../../ng-swagger-gen/models/addition-link";

@Component({
  selector: "hbr-artifact-dependencies",
  templateUrl: "./dependencies.component.html",
  styleUrls: ["./dependencies.component.scss"],
})
export class DependenciesComponent implements OnInit {
  @Input()
  dependenciesLink: AdditionLink;
  dependencyList: ArtifactDependency[] = [];
  constructor( private errorHandler: ErrorHandler,
               private additionsService: AdditionsService) {}

  ngOnInit(): void {
    this.getDependencyList();
  }
  getDependencyList() {
    if (this.dependenciesLink
        && !this.dependenciesLink.absolute
        && this.dependenciesLink.href) {
      this.additionsService.getDetailByLink(this.dependenciesLink.href).subscribe(
        res => {
          this.dependencyList = res;
        }, error => {
          this.errorHandler.error(error);
        }
      );
    }
  }
}
