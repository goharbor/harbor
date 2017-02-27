import { Component, OnInit, ViewChild } from '@angular/core';
import 'rxjs/add/operator/toPromise';

import { UserService } from './user.service';
import { User } from './user';
import { NewUserModalComponent } from './new-user-modal.component';

@Component({
  selector: 'harbor-user',
  templateUrl: 'user.component.html',
  styleUrls: ['user.component.css'],

  providers: [UserService]
})

export class UserComponent implements OnInit {
  users: User[] = [];
  originalUsers: Promise<User[]>;
  private onGoing: boolean = false;

  @ViewChild(NewUserModalComponent)
  private newUserDialog: NewUserModalComponent;

  constructor(private userService: UserService) { }

  private isMatchFilterTerm(terms: string, testedItem: string): boolean {
    return testedItem.indexOf(terms) != -1;
  }

  public get inProgress(): boolean {
    return this.onGoing;
  }

  ngOnInit(): void {
    this.refreshUser();
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

    //Value copy
    let updatedUser: User = Object.assign({}, user);

    if (updatedUser.has_admin_role === 0) {
      updatedUser.has_admin_role = 1;//Set as admin
    } else {
      updatedUser.has_admin_role = 0;//Set as none admin
    }

    this.userService.updateUser(updatedUser)
      .then(() => {
        //Change view now
        user.has_admin_role = updatedUser.has_admin_role;
      })
      .catch(error => console.error(error))//TODO:
  }

  //Delete the specified user
  deleteUser(userId: number): void {
    if (userId === 0) {
      return;
    }

    this.userService.deleteUser(userId)
      .then(() => {
        //Remove it from current user list
        //and then view refreshed
        this.originalUsers.then(users => {
          this.users = users.filter(user => user.user_id != userId);
        });
      })
      .catch(error => console.error(error));//TODO:
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
        console.error(error);//TODO:
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
