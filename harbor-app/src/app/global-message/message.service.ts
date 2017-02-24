import { Injectable } from '@angular/core';
import { Subject } from 'rxjs/Subject';

@Injectable()
export class MessageService {

  private messageAnnouncedSource = new Subject<string>();

  messageAnnounced$ = this.messageAnnouncedSource.asObservable();

  announceMessage(message: string) {
    this.messageAnnouncedSource.next(message);
  }
}