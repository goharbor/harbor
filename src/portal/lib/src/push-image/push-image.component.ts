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

  @ViewChild("tagCopy", { static: false }) tagCopyInput: CopyInputComponent;
  @ViewChild("pushCopy", { static: false }) pushCopyInput: CopyInputComponent;
  @ViewChild("copyAlert", { static: false }) copyAlert: InlineAlertComponent;

  public get tagCommand(): string {
    return `docker tag SOURCE_IMAGE[:TAG] ${this.registryUrl}/${
      this.projectName
    }/IMAGE[:TAG]`;
  }

  public get pushCommand(): string {
    return `docker push ${this.registryUrl}/${this.projectName}/IMAGE[:TAG]`;
  }

  onclick(): void {
    if (this.tagCopyInput) {
      this.tagCopyInput.reset();
    }

    if (this.pushCopyInput) {
      this.pushCopyInput.reset();
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
