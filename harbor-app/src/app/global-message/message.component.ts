import { Component, Input } from '@angular/core';

import { Message } from './message';
import { MessageService } from './message.service';

import { AlertType, dismissInterval } from '../shared/shared.const';

@Component({
  selector: 'global-message',
  templateUrl: 'message.component.html'
})
export class MessageComponent {
  
  globalMessageOpened: boolean;
  globalMessage: Message = new Message();
  
  constructor(messageService: MessageService) {
    messageService.messageAnnounced$.subscribe(
      message=>{ 
        this.globalMessageOpened = true;
        this.globalMessage = message;
        console.log('received message:' + message);
      }
    );
    // Make the message alert bar dismiss after several intervals.
    setInterval(()=>this.onClose(), dismissInterval);
  }
  
  onClose() {
    this.globalMessageOpened = false;
  }
}