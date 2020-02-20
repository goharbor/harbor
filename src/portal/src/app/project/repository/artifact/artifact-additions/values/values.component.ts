import {
  Component,
  Input,
  OnInit,
} from "@angular/core";
import { AdditionsService } from "../additions.service";
import { AdditionLink } from "../../../../../../../ng-swagger-gen/models/addition-link";
import { ErrorHandler } from "../../../../../../lib/utils/error-handler";


@Component({
  selector: "hbr-artifact-values",
  templateUrl: "./values.component.html",
  styleUrls: ["./values.component.scss"],
})
export class ValuesComponent implements OnInit {
  @Input()
  valuesLink: AdditionLink;

  values: any;

  // Default set to yaml file
  valueMode = true;
  valueHover = false;
  yamlHover = true;

  constructor(private errorHandler: ErrorHandler,
              private additionsService: AdditionsService) {
  }

  ngOnInit(): void {
    if (this.valuesLink && !this.valuesLink.absolute && this.valuesLink.href) {
      this.additionsService.getDetailByLink(this.valuesLink.href).subscribe(
        res => {
          this.values = res;
        }, error => {
          this.errorHandler.error(error);
        }
      );
    }
  }

  public get isValueMode() {
    return this.valueMode;
  }

  isHovering(view: string) {
    if (view === 'value') {
      return this.valueHover;
    } else {
      return this.yamlHover;
    }
  }

  showYamlFile(showYaml: boolean) {
    this.valueMode = !showYaml;
  }

  mouseEnter(mode: string) {
    if (mode === "value") {
      this.valueHover = true;
    } else {
      this.yamlHover = true;
    }
  }

  mouseLeave(mode: string) {
    if (mode === "value") {
      this.valueHover = false;
    } else {
      this.yamlHover = false;
    }
  }
}
