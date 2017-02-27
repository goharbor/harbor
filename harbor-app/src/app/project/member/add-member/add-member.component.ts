import { Component, Input, EventEmitter, Output } from '@angular/core';
import { Response } from '@angular/http';
import { MemberService } from '../member.service';
import { MessageService } from '../../../global-message/message.service';
import { AlertType } from '../../../shared/shared.const';


import { TranslateService } from '@ngx-translate/core';

import { Member } from '../member';

@Component({
  selector: 'add-member',
  templateUrl: 'add-member.component.html'
})
export class AddMemberComponent {

  member: Member = new Member();
  addMemberOpened: boolean;
  errorMessage: string;
  hasError: boolean;

  @Input() projectId: number;
  @Output() added = new EventEmitter<boolean>();

  constructor(private memberService: MemberService, 
              private messageService: MessageService, 
              private translateService: TranslateService) {}

  onSubmit(): void {
    this.hasError = false;
    console.log('Adding member:' + JSON.stringify(this.member));
    this.memberService
        .addMember(this.projectId, this.member.username, this.member.role_id)
        .subscribe(
          response=>{
            console.log('Added member successfully.');
            this.added.emit(true);
            this.addMemberOpened = false;
          },
          error=>{
            this.hasError = true;
            if (error instanceof Response) { 
            switch(error.status){
              case 404:
                this.translateService.get('MEMBER.USERNAME_DOES_NOT_EXISTS').subscribe(res=>this.errorMessage = res);
                break;
              case 409:
                this.translateService.get('MEMBER.USERNAME_ALREADY_EXISTS').subscribe(res=>this.errorMessage = res);
                break;
              default:
                this.translateService.get('MEMBER.UNKNOWN_ERROR').subscribe(res=>{
                  this.errorMessage = res;
                  this.messageService.announceMessage(error.status, this.errorMessage, AlertType.DANGER);
                });
                
              }
            }
            console.log('Failed to add member of project:' + this.projectId, ' with error:' + error);
          }
        );

    
  }

  openAddMemberModal(): void {
    this.hasError = false;
    this.member = new Member();
    this.addMemberOpened = true;
  }
}