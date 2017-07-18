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
import {
  Component,
  Output,
  EventEmitter,
  ViewChild,
  AfterViewChecked
} from '@angular/core';
import { NgForm } from '@angular/forms';

import { EndpointService } from '../service/endpoint.service';
import { ErrorHandler } from '../error-handler/index';
import { ActionType } from '../shared/shared.const';

import { InlineAlertComponent } from '../inline-alert/inline-alert.component';

import { Endpoint } from '../service/interface';

import { TranslateService } from '@ngx-translate/core';

import { CREATE_EDIT_ENDPOINT_STYLE } from './create-edit-endpoint.component.css';
import { CREATE_EDIT_ENDPOINT_TEMPLATE } from './create-edit-endpoint.component.html';


import { toPromise } from '../utils';

const FAKE_PASSWORD = 'rjGcfuRu';

@Component({
  selector: 'create-edit-endpoint',
  template: CREATE_EDIT_ENDPOINT_TEMPLATE,
  styles: [CREATE_EDIT_ENDPOINT_STYLE]
})
export class CreateEditEndpointComponent implements AfterViewChecked {

  modalTitle: string;
  createEditDestinationOpened: boolean;
  editable: boolean;
  testOngoing: boolean;

  actionType: ActionType;

  target: Endpoint = this.initEndpoint;
  initVal: Endpoint = this.initEndpoint;

  targetForm: NgForm;

  staticBackdrop: boolean = true;
  closable: boolean = false;

  @ViewChild('targetForm')
  currentForm: NgForm;

  hasChanged: boolean;
  endpointHasChanged: boolean;
  targetNameHasChanged: boolean;

  @ViewChild(InlineAlertComponent)
  inlineAlert: InlineAlertComponent;

  @Output() reload = new EventEmitter<boolean>();


  get initEndpoint(): Endpoint {
    return {
      endpoint: "",
      name: "",
      username: "",
      password: "",
      type: 0
    };
  }

  get hasConnectData():boolean{
    return !this.target.endpoint || !this.target.username || !this.target.password;
  }

  constructor(
    private endpointService: EndpointService,
    private errorHandler: ErrorHandler,
    private translateService: TranslateService) { }

  openCreateEditTarget(editable: boolean, targetId?: number | string) {

    this.target = this.initEndpoint;
    this.editable = editable;
    this.createEditDestinationOpened = true;
    this.hasChanged = false;
    this.endpointHasChanged = false;
    this.targetNameHasChanged = false;

    this.testOngoing = false;

    if (targetId) {
      this.actionType = ActionType.EDIT;
      this.translateService.get('DESTINATION.TITLE_EDIT').subscribe(res => this.modalTitle = res);
      toPromise<Endpoint>(this.endpointService
        .getEndpoint(targetId))
        .then(
        target => {
          this.target = target;
          this.initVal.name = this.target.name;
          this.initVal.endpoint = this.target.endpoint;
          this.initVal.username = this.target.username;
          this.initVal.password = FAKE_PASSWORD;
          this.target.password = this.initVal.password;
        })
        .catch(error => this.errorHandler.error(error));
    } else {
      this.actionType = ActionType.ADD_NEW;
      this.translateService.get('DESTINATION.TITLE_ADD').subscribe(res => this.modalTitle = res);
    }
  }

  testConnection() {
    let payload: Endpoint = this.initEndpoint;
    if (this.endpointHasChanged) {
      payload.endpoint = this.target.endpoint;
      payload.username = this.target.username;
      payload.password = this.target.password;
    } else {
      payload.id = this.target.id;
    }

    this.testOngoing = true;
    toPromise<Endpoint>(this.endpointService
      .pingEndpoint(payload))
      .then(
      response => {
        this.testOngoing = false;
        this.inlineAlert.showInlineSuccess({ message: "DESTINATION.TEST_CONNECTION_SUCCESS" });
      }).catch(
      error => {
        this.testOngoing = false;
        this.inlineAlert.showInlineError('DESTINATION.TEST_CONNECTION_FAILURE');
      });
  }

  changedTargetName($event: any) {
    if (this.editable) {
      this.targetNameHasChanged = true;
    }
  }

  clearPassword($event: any) {
    if (this.editable) {
      this.target.password = '';
      this.endpointHasChanged = true;
    }
  }

  onSubmit() {
    switch (this.actionType) {
      case ActionType.ADD_NEW:
        this.addEndpoint();
        break;
      case ActionType.EDIT:
        this.updateEndpoint();
        break;
    }
  }

  addEndpoint() {
    toPromise<number>(this.endpointService
      .createEndpoint(this.target))
      .then(
      response => {
        this.translateService.get('DESTINATION.CREATED_SUCCESS')
          .subscribe(res => this.errorHandler.info(res));
        this.createEditDestinationOpened = false;
        this.reload.emit(true);
      })
      .catch(
      error => {
        let errorMessageKey = this.handleErrorMessageKey(error.status);
        this.translateService
          .get(errorMessageKey)
          .subscribe(res => {
            this.inlineAlert.showInlineError(res);
          });
      }
      );
  }

  updateEndpoint() {
    if (!(this.targetNameHasChanged || this.endpointHasChanged)) {
      this.createEditDestinationOpened = false;
      return;
    }
    let payload: Endpoint = this.initEndpoint;
    if (this.targetNameHasChanged) {
      payload.name = this.target.name;
      delete payload.endpoint;
    }
    if (this.endpointHasChanged) {
      payload.endpoint = this.target.endpoint;
      payload.username = this.target.username;
      payload.password = this.target.password;
      delete payload.name;
    }

    if (!this.target.id) { return; }
    toPromise<number>(this.endpointService
      .updateEndpoint(this.target.id, payload))
      .then(
      response => {
        this.translateService.get('DESTINATION.UPDATED_SUCCESS')
          .subscribe(res => this.errorHandler.info(res));
        this.createEditDestinationOpened = false;
        this.reload.emit(true);
      })
      .catch(
      error => {
        let errorMessageKey = this.handleErrorMessageKey(error.status);
        this.translateService
          .get(errorMessageKey)
          .subscribe(res => {
            this.inlineAlert.showInlineError(res);
          });
      }
      );
  }

  handleErrorMessageKey(status: number): string {
    switch (status) {
      case 409: this
        return 'DESTINATION.CONFLICT_NAME';
      case 400:
        return 'DESTINATION.INVALID_NAME';
      default:
        return 'UNKNOWN_ERROR';
    }
  }

  onCancel() {
    if (this.hasChanged) {
      this.inlineAlert.showInlineConfirmation({ message: 'ALERT.FORM_CHANGE_CONFIRMATION' });
    } else {
      this.createEditDestinationOpened = false;
      if (this.targetForm)
        this.targetForm.reset();
    }
  }

  confirmCancel(confirmed: boolean) {
    this.createEditDestinationOpened = false;
    this.inlineAlert.close();
  }

  ngAfterViewChecked(): void {
    this.targetForm = this.currentForm;
    if (this.targetForm) {
      let comparison: { [key: string]: any } = {
        targetName: this.initVal.name,
        endpointUrl: this.initVal.endpoint,
        username: this.initVal.username,
        password: this.initVal.password
      };
      let self: CreateEditEndpointComponent | any = this;
      if (self) {
        self.targetForm.valueChanges.subscribe((data: any) => {
          for (let key in data) {
            let current = data[key];
            let origin: string = comparison[key];
            if (((this.actionType === ActionType.EDIT && this.editable && !current) || current) &&
              current !== origin) {
              this.hasChanged = true;
              break;
            } else {
              this.hasChanged = false;
              this.inlineAlert.close();
            }
          }
        });
      }
    }
  }

}