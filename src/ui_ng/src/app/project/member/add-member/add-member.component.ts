// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { Component, Input, EventEmitter, Output, ViewChild, AfterViewChecked } from '@angular/core';
import { Response } from '@angular/http';
import { NgForm } from '@angular/forms';
 
import { MemberService } from '../member.service';

import { MessageHandlerService } from '../../../shared/message-handler/message-handler.service';
import { InlineAlertComponent } from '../../../shared/inline-alert/inline-alert.component';

import { TranslateService } from '@ngx-translate/core';

import { Member } from '../member';

@Component({
  selector: 'add-member',
  templateUrl: 'add-member.component.html',
  styleUrls: [ 'add-member.component.css' ]
})
export class AddMemberComponent implements AfterViewChecked {

  member: Member = new Member();
  initVal: Member = new Member();

  addMemberOpened: boolean;
  
  memberForm: NgForm;

  staticBackdrop: boolean = true;
  closable: boolean = false;

  @ViewChild('memberForm')
  currentForm: NgForm;

  hasChanged: boolean;

  @ViewChild(InlineAlertComponent)
  inlineAlert: InlineAlertComponent;

  @Input() projectId: number;
  @Output() added = new EventEmitter<boolean>();

  constructor(private memberService: MemberService, 
              private messageHandlerService: MessageHandlerService, 
              private translateService: TranslateService) {}

  onSubmit(): void {
    if(!this.member.username || this.member.username.length === 0) { return; }
    this.memberService
        .addMember(this.projectId, this.member.username, +this.member.role_id)
        .subscribe(
          response=>{
            this.messageHandlerService.showSuccess('MEMBER.ADDED_SUCCESS');
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
              if(this.messageHandlerService.isAppLevel(error)) {
                this.messageHandlerService.handleError(error);
                this.addMemberOpened = false;
              } else {
               this.translateService
                  .get(errorMessageKey)
                  .subscribe(errorMessage=>this.inlineAlert.showInlineError(errorMessage));
              }
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
      this.memberForm.reset();
    }
  }

  ngAfterViewChecked(): void {
    this.memberForm = this.currentForm;
    if(this.memberForm) {
      this.memberForm.valueChanges.subscribe(data=>{
       for(let i in data) {
          let origin = this.initVal[i];          
          let current = data[i];
          if(current && current !== origin) {
            this.hasChanged = true;
            break;
          } else {
            this.hasChanged = false;
            this.inlineAlert.close();
          }
        }
      });
    }
  }

  confirmCancel(confirmed: boolean) {
    this.addMemberOpened = false;
    this.inlineAlert.close();
    this.memberForm.reset();
  }

  openAddMemberModal(): void {
    this.memberForm.reset();
    this.member = new Member();
    this.addMemberOpened = true;
    this.hasChanged = false;
    this.member.role_id = 1;
  }

}