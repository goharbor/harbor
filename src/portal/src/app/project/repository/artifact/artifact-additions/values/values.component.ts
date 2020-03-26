import {
  Component,
  Input,
  OnInit,
} from "@angular/core";
import { AdditionsService } from "../additions.service";
import { AdditionLink } from "../../../../../../../ng-swagger-gen/models/addition-link";
import { ErrorHandler } from "../../../../../../lib/utils/error-handler";
import  * as yaml  from "js-yaml";
import { finalize } from "rxjs/operators";
import { isObject } from "../../../../../../lib/utils/utils";

@Component({
  selector: "hbr-artifact-values",
  templateUrl: "./values.component.html",
  styleUrls: ["./values.component.scss"],
})
export class ValuesComponent implements OnInit {
  @Input()
  valuesLink: AdditionLink;

  values: string;
  valuesObj: object = {};

  // Default set to yaml file
  valueMode = false;
  valueHover = false;
  yamlHover = true;
  loading: boolean = false;
  constructor(private errorHandler: ErrorHandler,
              private additionsService: AdditionsService) {
  }

  ngOnInit(): void {
    if (this.valuesLink && !this.valuesLink.absolute && this.valuesLink.href) {
      this.loading = true;
      this.additionsService.getDetailByLink(this.valuesLink.href, true)
        .pipe(finalize(() => this.loading = false))
        .subscribe(
        res => {
          try {
            this.format(yaml.safeLoad(res));
            this.values = res;
          } catch (e) {
            this.errorHandler.error(e);
          }
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
  format(obj: object) {
      for (let name in obj) {
        if (obj.hasOwnProperty(name)) {
          if (isObject(obj[name])) {
            for (let key in obj[name]) {
              if (obj[name].hasOwnProperty(key)) {
                  obj[`${name}.${key}`] = obj[name][key];
              }
            }
            delete obj[name];
            this.format(obj);
          } else {
            this.valuesObj[name] = obj[name];
          }
        }
      }
  }
}
