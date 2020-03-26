
import { finalize } from 'rxjs/operators';
import { Subscription } from "rxjs";
import { Component, OnInit, EventEmitter, Output, ChangeDetectorRef, OnDestroy, ViewChild } from "@angular/core";
import { NgForm } from "@angular/forms";
import { GroupService } from "../group.service";
import { MessageHandlerService } from "./../../shared/message-handler/message-handler.service";
import { SessionService } from "./../../shared/session.service";
import { UserGroup } from "./../group";
import { AppConfigService } from "../../services/app-config.service";
import { GroupType } from "../../../lib/entities/shared.const";

@Component({
  selector: "hbr-add-group-modal",
  templateUrl: "./add-group-modal.component.html",
  styleUrls: ["./add-group-modal.component.scss"]
})
export class AddGroupModalComponent implements OnInit, OnDestroy {
  opened = false;
  mode = "create";
  dnTooltip = 'TOOLTIP.ITEM_REQUIRED';

  group: UserGroup;

  formChangeSubscription: Subscription;

  @ViewChild('groupForm', { static: true })
  groupForm: NgForm;

  submitted = false;

  @Output() dataChange = new EventEmitter();

  isLdapMode: boolean;
  isHttpAuthMode: boolean;
  isOidcMode: boolean;
  constructor(
    private session: SessionService,
    private msgHandler: MessageHandlerService,
    private appConfigService: AppConfigService,
    private groupService: GroupService,
    private cdr: ChangeDetectorRef
  ) { }

  ngOnInit() {
    if (this.appConfigService.isLdapMode()) {
      this.isLdapMode = true;
    }
    if (this.appConfigService.isHttpAuthMode()) {
      this.isHttpAuthMode = true;
    }
    if (this.appConfigService.isOidcMode()) {
      this.isOidcMode = true;
    }
    this.group = new UserGroup(this.isLdapMode ? GroupType.LDAP_TYPE : this.isHttpAuthMode ? GroupType.HTTP_TYPE : GroupType.OIDC_TYPE);
  }


  ngOnDestroy() { }

  public get isFormValid(): boolean {
    return this.groupForm.valid;
  }

  public open(group?: UserGroup, editMode: boolean = false): void {
    this.resetGroup();
    if (editMode) {
      this.mode = "edit";
      Object.assign(this.group, group);
    } else {
      this.mode = "create";
    }
    this.opened = true;
  }

  public close(): void {
    this.opened = false;
    this.resetGroup();
  }

  save(): void {
    if (this.mode === "create") {
      this.createGroup();
    } else {
      this.editGroup();
    }
  }

  createGroup() {
    let groupCopy = Object.assign({}, this.group);
    this.groupService
      .createGroup(groupCopy).pipe(
        finalize(() => this.close()))
      .subscribe(
        res => {
          this.msgHandler.showSuccess("GROUP.ADD_GROUP_SUCCESS");
          this.dataChange.emit();
        },
        error => this.msgHandler.handleError(error)
      );
  }

  editGroup() {
    let groupCopy = Object.assign({}, this.group);
    this.groupService
      .editGroup(groupCopy).pipe(
        finalize(() => this.close()))
      .subscribe(
        res => {
          this.msgHandler.showSuccess("GROUP.EDIT_GROUP_SUCCESS");
          this.dataChange.emit();
        },
        error => this.msgHandler.handleError(error)
      );
  }

  resetGroup() {
    this.group = new UserGroup(this.isLdapMode ? GroupType.LDAP_TYPE : this.isHttpAuthMode ? GroupType.HTTP_TYPE : GroupType.OIDC_TYPE);
    this.groupForm.reset();
  }
}
