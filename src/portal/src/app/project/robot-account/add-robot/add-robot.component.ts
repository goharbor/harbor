import {
  Component,
  OnInit,
  Input,
  ViewChild,
  OnDestroy,
  Output,
  EventEmitter,
  ChangeDetectorRef
} from "@angular/core";
import { Robot } from "../robot";
import { NgForm } from "@angular/forms";
import { Subject } from "rxjs";
import { debounceTime, finalize } from "rxjs/operators";
import { RobotService } from "../robot-account.service";
import { TranslateService } from "@ngx-translate/core";
import { ErrorHandler } from "@harbor/ui";
import { MessageHandlerService } from "../../../shared/message-handler/message-handler.service";
import { InlineAlertComponent } from "../../../shared/inline-alert/inline-alert.component";

@Component({
  selector: "add-robot",
  templateUrl: "./add-robot.component.html",
  styleUrls: ["./add-robot.component.scss"]
})
export class AddRobotComponent implements OnInit, OnDestroy {
  addRobotOpened: boolean;
  copyToken: boolean;
  robotToken: string;
  robotAccount: string;
  isSubmitOnGoing = false;
  closable: boolean = false;
  staticBackdrop: boolean = true;
  createSuccess: string;
  isRobotNameValid: boolean = true;
  checkOnGoing: boolean = false;
  robot: Robot = new Robot();
  robotNameChecker: Subject<string> = new Subject<string>();
  nameTooltipText = "ROBOT_ACCOUNT.ROBOT_NAME";
  robotForm: NgForm;
  @Input() projectId: number;
  @Input() projectName: string;
  @Output() create = new EventEmitter<boolean>();
  @ViewChild("robotForm") currentForm: NgForm;
  @ViewChild("copyAlert") copyAlert: InlineAlertComponent;
  constructor(
    private robotService: RobotService,
    private translate: TranslateService,
    private errorHandler: ErrorHandler,
    private cdr: ChangeDetectorRef,
    private messageHandlerService: MessageHandlerService
  ) {}

  ngOnInit(): void {
    this.robotNameChecker.pipe(debounceTime(800)).subscribe((name: string) => {
      let cont = this.currentForm.controls["robot_name"];
      if (cont) {
        this.isRobotNameValid = cont.valid;
        if (this.isRobotNameValid) {
          this.checkOnGoing = true;
          this.robotService
            .listRobotAccount(this.projectId)
            .pipe(
              finalize(() => {
                this.checkOnGoing = false;
                let hnd = setInterval(() => this.cdr.markForCheck(), 100);
                setTimeout(() => clearInterval(hnd), 2000);
              })
            )
            .subscribe(
              response => {
                if (response && response.length) {
                  if (
                    response.find(target => {
                      return target.name === "robot$" + cont.value;
                    })
                  ) {
                    this.isRobotNameValid = false;
                    this.nameTooltipText = "ROBOT_ACCOUNT.ACCOUNT_EXISTING";
                  }
                }
              },
              error => {
                this.errorHandler.error(error);
              }
            );
        } else {
          this.nameTooltipText = "ROBOT_ACCOUNT.ROBOT_NAME";
        }
      }
    });
  }

  openAddRobotModal(): void {
    if (this.isSubmitOnGoing) {
      return;
    }
    this.robot.name = "";
    this.robot.description = "";
    this.addRobotOpened = true;
    this.isRobotNameValid = true;
    this.robot = new Robot();
    this.nameTooltipText = "ROBOT_ACCOUNT.ROBOT_NAME";
    this.copyAlert.close();
  }

  onCancel(): void {
    this.addRobotOpened = false;
  }

  ngOnDestroy(): void {
    this.robotNameChecker.unsubscribe();
  }

  onSubmit(): void {
    if (this.isSubmitOnGoing) {
      return;
    }
    this.isSubmitOnGoing = true;
    this.robotService
      .addRobotAccount(
        this.projectId,
        this.robot,
        this.projectName
      )
      .subscribe(
        response => {
          this.isSubmitOnGoing = false;
          this.robotToken = response.token;
          this.robotAccount = response.name;
          this.copyToken = true;
          this.create.emit(true);
          this.translate
            .get("ROBOT_ACCOUNT.CREATED_SUCCESS", { param: this.robotAccount })
            .subscribe((res: string) => {
              this.createSuccess = res;
            });
          this.addRobotOpened = false;
        },
        error => {
          this.isSubmitOnGoing = false;
          this.copyAlert.showInlineError(error);
        }
      );
  }

  isValid(): boolean {
    return (
      this.currentForm &&
      this.currentForm.valid &&
      !this.isSubmitOnGoing &&
      this.isRobotNameValid &&
      !this.checkOnGoing
    );
  }
  get shouldDisable(): boolean {
    if (this.robot && this.robot.access) {
      return (
        !this.isValid() ||
        (!this.robot.access.isPushOrPullImage && !this.robot.access.isPullImage
          && !this.robot.access.isPullChart && !this.robot.access.isPushChart)
      );
    }
  }

  // Handle the form validation
  handleValidation(): void {
    let cont = this.currentForm.controls["robot_name"];
    if (cont) {
      this.robotNameChecker.next(cont.value);
    }
  }

  onCpError($event: any): void {
    if (this.copyAlert) {
      this.copyAlert.showInlineError("PUSH_IMAGE.COPY_ERROR");
    }
  }

  onCpSuccess($event: any): void {
    this.copyToken = false;
    this.translate
      .get("ROBOT_ACCOUNT.COPY_SUCCESS", { param: this.robotAccount })
      .subscribe((res: string) => {
        this.messageHandlerService.showSuccess(res);
      });
  }

  radioChange(radioId: string ): void {
    if (radioId === 'image-permission-push') {
      this.robot.access.isPullImage = true;
      this.robot.access.isPushOrPullImage = false;
    } else if (radioId === 'image-permission-push-and-pull') {
      this.robot.access.isPullImage = false;
      this.robot.access.isPushOrPullImage = true;
    }
  }
}
