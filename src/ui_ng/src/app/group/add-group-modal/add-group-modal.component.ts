import { Subscription } from 'rxjs/Subscription';
import { Component, OnInit, EventEmitter, Output, ChangeDetectorRef, OnDestroy, ViewChild } from "@angular/core";
import { NgForm } from "@angular/forms";
import "rxjs/add/operator/finally";

import { GroupService } from "../group.service";
import { MessageHandlerService } from "./../../shared/message-handler/message-handler.service";
import { SessionService } from "./../../shared/session.service";
import { UserGroup } from "./../group";

@Component({
  selector: "hbr-add-group-modal",
  templateUrl: "./add-group-modal.component.html",
  styleUrls: ["./add-group-modal.component.scss"]
})
export class AddGroupModalComponent implements OnInit, OnDestroy {
  opened = false;
  mode = "create";
  dnTooltip = 'TOOLTIP.ITEM_REQUIRED';

  group: UserGroup = new UserGroup();

  formChangeSubscription: Subscription;

  @ViewChild('groupForm')
  groupForm: NgForm;

  submitted = false;

  @Output() dataChange = new EventEmitter();

  constructor(
    private session: SessionService,
    private msgHandler: MessageHandlerService,
    private groupService: GroupService,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit() { }


  ngOnDestroy() { }

  public get isDNInvalid(): boolean {
    let dnControl = this.groupForm.controls['ldap_group_dn'];
    return  dnControl && dnControl.invalid && (dnControl.dirty || dnControl.touched);
  }

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
      .createGroup(groupCopy)
      .finally(() => this.close())
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
      .editGroup(groupCopy)
      .finally(() => this.close())
      .subscribe(
        res => {
          this.msgHandler.showSuccess("ADD_GROUP_FAILURE");
          this.dataChange.emit();
        },
        error => this.msgHandler.handleError(error)
      );
  }

  resetGroup() {
    this.group = new UserGroup();
    this.groupForm.reset();
  }
}
