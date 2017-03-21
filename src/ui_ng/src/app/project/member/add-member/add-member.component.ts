import { Component, Input, EventEmitter, Output, ViewChild, AfterViewChecked } from '@angular/core';
import { Response } from '@angular/http';
import { NgForm } from '@angular/forms';
 
import { MemberService } from '../member.service';
import { MessageService } from '../../../global-message/message.service';
import { AlertType } from '../../../shared/shared.const';


import { InlineAlertComponent } from '../../../shared/inline-alert/inline-alert.component';

import { TranslateService } from '@ngx-translate/core';

import { Member } from '../member';

@Component({
  selector: 'add-member',
  templateUrl: 'add-member.component.html'
})
export class AddMemberComponent implements AfterViewChecked {

  member: Member = new Member();
  addMemberOpened: boolean;
  
  memberForm: NgForm;

  @ViewChild('memberForm')
  currentForm: NgForm;

  hasChanged: boolean;

  @ViewChild(InlineAlertComponent)
  inlineAlert: InlineAlertComponent;

  @Input() projectId: number;
  @Output() added = new EventEmitter<boolean>();

  constructor(private memberService: MemberService, 
              private messageService: MessageService, 
              private translateService: TranslateService) {}

  onSubmit(): void {
    console.log('Adding member:' + JSON.stringify(this.member));
    this.memberService
        .addMember(this.projectId, this.member.username, +this.member.role_id)
        .subscribe(
          response=>{
            console.log('Added member successfully.');
            this.added.emit(true);
            this.addMemberOpened = false;
          },
          error=>{
            if (error instanceof Response) {             
            let errorMessageKey: string;
            switch(error.status){
              case 404:
                errorMessageKey = 'MEMBER.USERNAME_DOES_NOT_EXISTS';
                break;
              case 409:
                errorMessageKey = 'MEMBER.USERNAME_ALREADY_EXISTS';
                break;
              default:
                errorMessageKey = 'MEMBER.UNKNOWN_ERROR';              
              }
               this.translateService
                  .get(errorMessageKey)
                  .subscribe(errorMessage=>this.inlineAlert.showInlineError(errorMessage));
            }
            console.log('Failed to add member of project:' + this.projectId, ' with error:' + error);
          }
        );
  }

  onCancel() {
    if(this.hasChanged) {
      this.inlineAlert.showInlineConfirmation({message: 'ALERT.FORM_CHANGE_CONFIRMATION'});
    } else {
      this.addMemberOpened = false;
    }
  }

  ngAfterViewChecked(): void {
    this.memberForm = this.currentForm;
    if(this.memberForm) {
      this.memberForm.valueChanges.subscribe(data=>{
        for(let i in data) {
          let item = data[i];
          if(typeof item === 'string' && (<string>item).trim().length !== 0) {
            this.hasChanged = true;
            break;
          } else if (typeof item === 'boolean' && (<boolean>item)) {
            this.hasChanged = true;
            break;
          } else if (typeof item === 'number' && (<number>item) !== 0) {
            this.hasChanged = true;
            break;
          } else {
            this.hasChanged = false;
            this.inlineAlert.close();
            break;
          }
        }
      });
    }
  }

  confirmCancel(confirmed: boolean) {
    this.addMemberOpened = false;
    this.inlineAlert.close();
  }

  openAddMemberModal(): void {
    this.member = new Member();
    this.addMemberOpened = true;
    this.hasChanged = false;
  }

}