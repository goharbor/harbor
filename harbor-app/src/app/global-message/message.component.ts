import { Component, Input, OnInit } from '@angular/core';
import { Router } from '@angular/router';

import { TranslateService } from '@ngx-translate/core';

import { Message } from './message';
import { MessageService } from './message.service';

import { AlertType, dismissInterval, httpStatusCode, CommonRoutes } from '../shared/shared.const';

@Component({
  selector: 'global-message',
  templateUrl: 'message.component.html'
})
export class MessageComponent implements OnInit {

  @Input() isAppLevel: boolean;
  globalMessage: Message = new Message();
  globalMessageOpened: boolean;
  messageText: string = "";
  private timer: any = null;

  constructor(
    private messageService: MessageService,
    private router: Router,
    private translate: TranslateService) { }

  ngOnInit(): void {
    //Only subscribe application level message
    if (this.isAppLevel) {
      this.messageService.appLevelAnnounced$.subscribe(
        message => {
          this.globalMessageOpened = true;
          this.globalMessage = message;
          this.messageText = message.message;

          this.translateMessage(message);
        }
      )
    } else {
      //Only subscribe general messages
      this.messageService.messageAnnounced$.subscribe(
        message => {
          this.globalMessageOpened = true;
          this.globalMessage = message;
          this.messageText = message.message;

          this.translateMessage(message);

          // Make the message alert bar dismiss after several intervals.
          //Only for this case
          this.timer = setTimeout(() => this.onClose(), dismissInterval);
        }
      );
    }
  }

  //Translate or refactor the message shown to user
  translateMessage(msg: Message): void {
    let key = "UNKNOWN_ERROR", param = "";
    if (msg && msg.message) {
      key = (typeof msg.message === "string" ? msg.message.trim() : msg.message);
      if (key === "") {
        key = "UNKNOWN_ERROR";
      }
    }

    //Override key for HTTP 401 and 403
    if (this.globalMessage.statusCode === httpStatusCode.Unauthorized) {
      key = "UNAUTHORIZED_ERROR";
    } else if (this.globalMessage.statusCode === httpStatusCode.Forbidden) {
      key = "FORBIDDEN_ERROR";
    } 

    this.translate.get(key, { 'param': param }).subscribe((res: string) => this.messageText = res);
  }

  public get needAuth(): boolean {
    return this.globalMessage ?
      (this.globalMessage.statusCode === httpStatusCode.Unauthorized) ||
      (this.globalMessage.statusCode === httpStatusCode.Forbidden) : false;
  }

  //Show message text
  public get message(): string {
    return this.messageText;
  }

  signIn(): void {
    this.router.navigateByUrl(CommonRoutes.EMBEDDED_SIGN_IN);
  }

  onClose() {
    if (this.timer) {
      clearTimeout(this.timer);
    }
    this.globalMessageOpened = false;
  }
}