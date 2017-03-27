import { Component, OnInit, ViewChild, OnDestroy } from '@angular/core';
import 'rxjs/add/operator/toPromise';
import { Subscription } from 'rxjs/Subscription';

import { UserService } from './user.service';
import { User } from './user';
import { NewUserModalComponent } from './new-user-modal.component';
import { TranslateService } from '@ngx-translate/core';
import { ConfirmationDialogService } from '../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../shared/confirmation-dialog/confirmation-message';
import { ConfirmationState, ConfirmationTargets, AlertType, httpStatusCode } from '../shared/shared.const'
import { errorHandler, accessErrorHandler } from '../shared/shared.utils';
import { MessageService } from '../global-message/message.service';

import { SessionService } from '../shared/session.service';

@Component({
  selector: 'harbor-user',
  templateUrl: 'user.component.html',
  styleUrls: ['user.component.css'],

  providers: [UserService]
})

export class UserComponent implements OnInit, OnDestroy {
  users: User[] = [];
  originalUsers: Promise<User[]>;
  private onGoing: boolean = false;
  private adminMenuText: string = "";
  private adminColumn: string = "";
  private deletionSubscription: Subscription;

  @ViewChild(NewUserModalComponent)
  private newUserDialog: NewUserModalComponent;

  constructor(
    private userService: UserService,
    private translate: TranslateService,
    private deletionDialogService: ConfirmationDialogService,
    private msgService: MessageService,
    private session: SessionService) {
    this.deletionSubscription = deletionDialogService.confirmationConfirm$.subscribe(confirmed => {
      if (confirmed &&
        confirmed.source === ConfirmationTargets.USER &&
        confirmed.state === ConfirmationState.CONFIRMED) {
        this.delUser(confirmed.data);
      }
    });
  }

  private isMySelf(uid: number): boolean {
    let currentUser = this.session.getCurrentUser();
    if(currentUser){
      if(currentUser.user_id === uid ){
        return true;
      }
    }

    return false;
  }

  private isMatchFilterTerm(terms: string, testedItem: string): boolean {
    return testedItem.indexOf(terms) != -1;
  }

  isSystemAdmin(u: User): string {
    if (!u) {
      return "{{MISS}}";
    }
    let key: string = u.has_admin_role ? "USER.IS_ADMIN" : "USER.IS_NOT_ADMIN";
    this.translate.get(key).subscribe((res: string) => this.adminColumn = res);
    return this.adminColumn;
  }

  adminActions(u: User): string {
    if (!u) {
      return "{{MISS}}";
    }
    let key: string = u.has_admin_role ? "USER.DISABLE_ADMIN_ACTION" : "USER.ENABLE_ADMIN_ACTION";
    this.translate.get(key).subscribe((res: string) => this.adminMenuText = res);
    return this.adminMenuText;
  }

  public get inProgress(): boolean {
    return this.onGoing;
  }

  ngOnInit(): void {
    this.refreshUser();
  }

  ngOnDestroy(): void {
    if (this.deletionSubscription) {
      this.deletionSubscription.unsubscribe();
    }
  }

  //Filter items by keywords
  doFilter(terms: string): void {
    this.originalUsers.then(users => {
      if (terms.trim() === "") {
        this.users = users;
      } else {
        this.users = users.filter(user => {
          return this.isMatchFilterTerm(terms, user.username);
        })
      }
    });
  }

  //Disable the admin role for the specified user
  changeAdminRole(user: User): void {
    //Double confirm user is existing
    if (!user || user.user_id === 0) {
      return;
    }

    if(this.isMySelf(user.user_id)){
      return;
    }

    //Value copy
    let updatedUser: User = {
      user_id: user.user_id
    };

    if (user.has_admin_role === 0) {
      updatedUser.has_admin_role = 1;//Set as admin
    } else {
      updatedUser.has_admin_role = 0;//Set as none admin
    }

    this.userService.updateUserRole(updatedUser)
      .then(() => {
        //Change view now
        user.has_admin_role = updatedUser.has_admin_role;
      })
      .catch(error => {
        if (!accessErrorHandler(error, this.msgService)) {
          this.msgService.announceMessage(500, errorHandler(error), AlertType.DANGER);
        }
      })
  }

  //Delete the specified user
  deleteUser(user: User): void {
    if (!user) {
      return;
    }

    if(this.isMySelf(user.user_id)){
      return; //Double confirm
    }

    //Confirm deletion
    let msg: ConfirmationMessage = new ConfirmationMessage(
      "USER.DELETION_TITLE",
      "USER.DELETION_SUMMARY",
      user.username,
      user,
      ConfirmationTargets.USER
    );
    this.deletionDialogService.openComfirmDialog(msg);
  }

  private delUser(user: User): void {
    this.userService.deleteUser(user.user_id)
      .then(() => {
        //Remove it from current user list
        //and then view refreshed
        this.originalUsers.then(users => {
          this.users = users.filter(u => u.user_id != user.user_id);
          this.msgService.announceMessage(500, "USER.DELETE_SUCCESS", AlertType.SUCCESS);
        });
      })
      .catch(error => {
        if (!accessErrorHandler(error, this.msgService)) {
          this.msgService.announceMessage(500, errorHandler(error), AlertType.DANGER);
        }
      });
  }

  //Refresh the user list
  refreshUser(): void {
    //Start to get
    this.onGoing = true;

    this.originalUsers = this.userService.getUsers()
      .then(users => {
        this.onGoing = false;

        this.users = users;
        return users;
      })
      .catch(error => {
        this.onGoing = false;
        if (!accessErrorHandler(error, this.msgService)) {
          this.msgService.announceMessage(500, errorHandler(error), AlertType.DANGER);
        }
      });
  }

  //Add new user
  addNewUser(): void {
    this.newUserDialog.open();
  }

  //Add user to the user list
  addUserToList(user: User): void {
    //Currently we can only add it by reloading all
    this.refreshUser();
  }

}
