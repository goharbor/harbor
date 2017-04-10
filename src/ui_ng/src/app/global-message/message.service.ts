import { Injectable } from '@angular/core';
import { Subject } from 'rxjs/Subject';
import { Message } from './message';
import { AlertType } from '../shared/shared.const';

@Injectable()
export class MessageService {

  messageAnnouncedSource = new Subject<Message>();
  appLevelAnnouncedSource = new Subject<Message>();
  clearSource = new Subject<boolean>();

  messageAnnounced$ = this.messageAnnouncedSource.asObservable();
  appLevelAnnounced$ = this.appLevelAnnouncedSource.asObservable();
  clearChan$ = this.clearSource.asObservable();
 
  announceMessage(statusCode: number, message: string, alertType: AlertType) {
    this.messageAnnouncedSource.next(Message.newMessage(statusCode, message, alertType));
  }

  announceAppLevelMessage(statusCode: number, message: string, alertType: AlertType) {
    this.appLevelAnnouncedSource.next(Message.newMessage(statusCode, message, alertType));
  }

  clear() {
    this.clearSource.next(true);
  }
}