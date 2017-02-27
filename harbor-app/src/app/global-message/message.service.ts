import { Injectable } from '@angular/core';
import { Subject } from 'rxjs/Subject';
import { Message } from './message';
import { AlertType } from '../shared/shared.const';

@Injectable()
export class MessageService {

  private messageAnnouncedSource = new Subject<Message>();

  messageAnnounced$ = this.messageAnnouncedSource.asObservable();
 
  announceMessage(statusCode: number, message: string, alertType: AlertType, isAppLevel?: boolean) {
    this.messageAnnouncedSource.next(Message.newMessage(statusCode, message, alertType, (isAppLevel) ? isAppLevel : false));
  }
}