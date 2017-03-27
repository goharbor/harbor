import { Injectable } from '@angular/core'
import { Subject } from 'rxjs/Subject';

import { MessageService } from '../../global-message/message.service';
import { AlertType, httpStatusCode } from '../../shared/shared.const';

@Injectable()
export class MessageHandlerService {

    constructor(private msgService: MessageService) { }

    //Handle the error and map it to the suitable message
    //base on the status code of error.

    public handleError(error: any | string): void {
        if (!error) {
            return;
        }
        console.log(JSON.stringify(error));

        if (!(error.statusCode || error.status)) {
            //treat as string message
            let msg = '' + error;
            this.msgService.announceMessage(500, msg, AlertType.DANGER);
        } else {
            let msg = 'UNKNOWN_ERROR';
            switch (error.statusCode || error.status) {
                case 400:
                    msg = "BAD_REQUEST_ERROR";
                    break;
                case 401:
                    msg = "UNAUTHORIZED_ERROR";
                    this.msgService.announceAppLevelMessage(error.statusCode, msg, AlertType.DANGER);
                    return;
                case 403:
                    msg = "FORBIDDEN_ERROR";
                    break;
                case 404:
                    msg = "NOT_FOUND_ERROR";
                    break;
                case 412:
                case 409:
                    msg = "CONFLICT_ERROR";
                    break;
                case 500:
                    msg = "SERVER_ERROR";
                    break;
                default:
                    break;
            }
            this.msgService.announceMessage(error.statusCode, msg, AlertType.DANGER);
        }
    }

    public showSuccess(message: string): void {
        if (message && message.trim() != "") {
            this.msgService.announceMessage(200, message, AlertType.SUCCESS);
        }
    }

    public showInfo(message: string): void {
        if (message && message.trim() != "") {
            this.msgService.announceMessage(200, message, AlertType.INFO);
        }
    }

    public showWarning(message: string): void {
        if (message && message.trim() != "") {
            this.msgService.announceMessage(400, message, AlertType.WARNING);
        }
    }

    public clear(): void {
        this.msgService.clear();
    }

    public isAppLevel(error): boolean {
        return error && error.statusCode === httpStatusCode.Unauthorized;
    }
}