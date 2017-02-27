import { Component, Input, EventEmitter, Output } from '@angular/core';
import { Response } from '@angular/http';
import { MemberService } from '../member.service';
import { MessageService } from '../../../global-message/message.service';
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

  constructor(private memberService: MemberService, private messageService: MessageService) {}

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
                this.errorMessage = 'Username does not exist.';
                break;
              case 409:
                this.errorMessage = 'Username already exists.';
                break;
              default:
                this.errorMessage = 'Unknow error occurred while adding member.';
                this.messageService.announceMessage(this.errorMessage);
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