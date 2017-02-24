import { Component } from '@angular/core';
import { MessageService } from './message.service';

@Component({
  selector: 'global-message',
  templateUrl: 'message.component.html'
})
export class MessageComponent {
  
  globalMessageOpened: boolean;
  globalMessage: string;

  constructor(messageService: MessageService) {
    messageService.messageAnnounced$.subscribe(
      message=>{
        this.globalMessageOpened = true;
        this.globalMessage = message;
        console.log('received message:' + message);
      }
    )
  }

  onClose() {
    this.globalMessageOpened = false;
  }
}