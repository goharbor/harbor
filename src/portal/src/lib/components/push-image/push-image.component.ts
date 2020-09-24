import { Component, Input, ViewChild } from "@angular/core";
import { CopyInputComponent } from "./copy-input.component";
import { InlineAlertComponent } from "../inline-alert/inline-alert.component";

@Component({
  selector: "hbr-push-image-button",
  templateUrl: "./push-image.component.html",
  styleUrls: ["./push-image.scss"],

  providers: []
})
export class PushImageButtonComponent {
  @Input() registryUrl: string = "unknown";
  @Input() projectName: string = "unknown";

  @ViewChild("tagCopyImage", {static: false}) tagCopyImageInput: CopyInputComponent;
  @ViewChild("pushCopyImage", {static: false}) pushCopImageyInput: CopyInputComponent;
  @ViewChild("tagCopyChart", {static: false}) tagCopyChartInput: CopyInputComponent;
  @ViewChild("pushCopyChart", {static: false}) pushCopyChartInput: CopyInputComponent;
  @ViewChild("pushCopyCnab", {static: false}) pushCopCnabyInput: CopyInputComponent;
  @ViewChild("copyAlert", {static: false}) copyAlert: InlineAlertComponent;

  public get tagCommandImage(): string {
    return `docker tag SOURCE_IMAGE[:TAG] ${this.registryUrl}/${
      this.projectName
    }/REPOSITORY[:TAG]`;
  }

  public get pushCommandImage(): string {
    return `docker push ${this.registryUrl}/${this.projectName}/REPOSITORY[:TAG]`;
  }
  public get tagCommandChart(): string {
    return `helm chart save CHART_PATH ${this.registryUrl}/${
      this.projectName
    }/REPOSITORY[:TAG]`;
  }

  public get pushCommandChart(): string {
    return `helm chart push ${this.registryUrl}/${this.projectName}/REPOSITORY[:TAG]`;
  }

  public get pushCommandCnab(): string {
    return `cnab-to-oci push CNAB_PATH --target ${this.registryUrl}/${this.projectName}/REPOSITORY[:TAG] --auto-update-bundle`;
  }

  onclick(): void {
    if (this.tagCopyImageInput) {
      this.tagCopyImageInput.reset();
    }

    if (this.pushCopImageyInput) {
      this.pushCopImageyInput.reset();
    }
    if (this.tagCopyChartInput) {
      this.tagCopyChartInput.reset();
    }

    if (this.pushCopyChartInput) {
      this.pushCopyChartInput.reset();
    }

    if (this.pushCopCnabyInput) {
      this.pushCopCnabyInput.reset();
    }

    if (this.copyAlert) {
      this.copyAlert.close();
    }
  }

  onCpError($event: any): void {
    if (this.copyAlert) {
      this.copyAlert.showInlineError("PUSH_IMAGE.COPY_ERROR");
    }
  }
}
