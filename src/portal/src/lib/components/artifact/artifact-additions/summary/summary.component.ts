import {
  Component,
  OnInit,
  Input
} from "@angular/core";
import { ErrorHandler } from "../../../../utils/error-handler";
import { AdditionsService } from "../additions.service";
import { AdditionLink } from "../../../../../../ng-swagger-gen/models/addition-link";

@Component({
  selector: "hbr-artifact-summary",
  templateUrl: "./summary.component.html",
  styleUrls: ["./summary.component.scss"],
})
export class SummaryComponent implements OnInit {
  @Input() summaryLink: AdditionLink;
  readme: string;
  constructor(
    private errorHandler: ErrorHandler,
    private additionsService: AdditionsService
  ) {}

  ngOnInit(): void {
    this.getReadme();
  }
  getReadme() {
    if (this.summaryLink
      && !this.summaryLink.absolute
      && this.summaryLink.href) {
      this.additionsService.getDetailByLink(this.summaryLink.href).subscribe(
        res => {
          this.readme = res;
        }, error => {
          this.errorHandler.error(error);
        }
      );
    }
  }
}
