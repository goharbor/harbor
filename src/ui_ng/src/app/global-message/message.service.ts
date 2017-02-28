import { Injectable } from '@angular/core';
import { Subject } from 'rxjs/Subject';
import { Message } from './message';
import { AlertType } from '../shared/shared.const';

@Injectable()
export class MessageService {

  private messageAnnouncedSource = new Subject<Message>();
  private appLevelAnnouncedSource = new Subject<Message>();

  messageAnnounced$ = this.messageAnnouncedSource.asObservable();
  appLevelAnnounced$ = this.appLevelAnnouncedSource.asObservable();
 
  announceMessage(statusCode: number, message: string, alertType: AlertType) {
    this.messageAnnouncedSource.next(Message.newMessage(statusCode, message, alertType));
  }

  announceAppLevelMessage(statusCode: number, message: string, alertType: AlertType) {
    this.appLevelAnnouncedSource.next(Message.newMessage(statusCode, message, alertType));
  }
}