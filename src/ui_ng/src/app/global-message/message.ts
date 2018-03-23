// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { AlertType } from '../shared/shared.const';

export class Message {
  statusCode: number;
  message: string;
  alertType: AlertType;
  isAppLevel: boolean = false;

  get type(): string {
    switch (this.alertType) {
      case AlertType.DANGER:
        return 'alert-danger';
      case AlertType.INFO:
        return 'alert-info';
      case AlertType.SUCCESS:
        return 'alert-success';
      case AlertType.WARNING:
        return 'alert-warning';
      default:
        return 'alert-warning';
    }
  }

  constructor() { }

  static newMessage(statusCode: number, message: string, alertType: AlertType): Message {
    let m = new Message();
    m.statusCode = statusCode;
    m.message = message;
    m.alertType = alertType;
    return m;
  }


  toString(): string {
    return 'Message with statusCode:' + this.statusCode +
      ', message:' + this.message +
      ', alert type:' + this.type;
  }
}