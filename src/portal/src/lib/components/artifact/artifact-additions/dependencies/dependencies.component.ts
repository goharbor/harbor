import {
  Component,
  OnInit,
  Input,
} from "@angular/core";
import { ArtifactDependency } from "../models";
import { ErrorHandler } from "../../../../utils/error-handler";
import { AdditionsService } from "../additions.service";

@Component({
  selector: "hbr-artifact-dependencies",
  templateUrl: "./dependencies.component.html",
  styleUrls: ["./dependencies.component.scss"],
})
export class DependenciesComponent implements OnInit {
  @Input()
  dependenciesLink: string;

  dependencyList: ArtifactDependency[] = [
    {
      "name": "redis",
      "version": "3.2.5",
      "repository": "https://kubernetes-charts.storage.googleapis.com"
    }
  ];
  constructor( private errorHandler: ErrorHandler,
               private additionsService: AdditionsService) {}

  ngOnInit(): void {
    if (this.dependenciesLink) {
      this.additionsService.getDetailByLink(this.dependenciesLink).subscribe(
        res => {
          this.dependencyList = res;
        }, error => {
          this.errorHandler.error(error);
        }
      );
    }
  }

}
