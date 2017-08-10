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
    AfterViewChecked,
    ChangeDetectorRef,
    OnDestroy
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

import { toPromise, clone, compareValue } from '../utils';

import { Subscription } from 'rxjs/Subscription';

const FAKE_PASSWORD = 'rjGcfuRu';

@Component({
    selector: 'create-edit-endpoint',
    template: CREATE_EDIT_ENDPOINT_TEMPLATE,
    styles: [CREATE_EDIT_ENDPOINT_STYLE]
})
export class CreateEditEndpointComponent implements AfterViewChecked, OnDestroy {
    modalTitle: string;
    createEditDestinationOpened: boolean;
    staticBackdrop: boolean = true;
    closable: boolean = false;

    actionType: ActionType;
    editable: boolean;

    target: Endpoint = this.initEndpoint();
    initVal: Endpoint;

    targetForm: NgForm;
    @ViewChild('targetForm')
    currentForm: NgForm;

    endpointHasChanged: boolean;
    targetNameHasChanged: boolean;

    testOngoing: boolean;
    onGoing: boolean;

    @ViewChild(InlineAlertComponent)
    inlineAlert: InlineAlertComponent;

    @Output() reload = new EventEmitter<boolean>();

    timerHandler: any;
    valueChangesSub: Subscription;
    formValues: { [key: string]: string } | any;

    constructor(
        private endpointService: EndpointService,
        private errorHandler: ErrorHandler,
        private translateService: TranslateService,
        private ref: ChangeDetectorRef
    ) { }

    public get hasChanged(): boolean {
        if (this.actionType === ActionType.ADD_NEW) {
            //Create new
            return this.target && (
                (this.target.endpoint && this.target.endpoint.trim() !== "") ||
                (this.target.name && this.target.name.trim() !== "") ||
                (this.target.username && this.target.username.trim() !== "") ||
                (this.target.password && this.target.password.trim() !== ""));
        } else {
            //Edit
            return !compareValue(this.target, this.initVal);
        }
    }

    public get isValid(): boolean {
        return !this.testOngoing &&
            !this.onGoing &&
            this.targetForm &&
            this.targetForm.valid &&
            this.editable &&
            (this.targetNameHasChanged || this.endpointHasChanged);
    }

    public get inProgress(): boolean {
        return this.onGoing || this.testOngoing;
    }

    ngOnDestroy(): void {
        if (this.valueChangesSub) {
            this.valueChangesSub.unsubscribe();
        }
    }


    initEndpoint(): Endpoint {
        return {
            endpoint: "",
            name: "",
            username: "",
            password: "",
            type: 0
        };
    }

    open(): void {
        this.createEditDestinationOpened = true;
    }

    close(): void {
        this.createEditDestinationOpened = false;
    }

    reset(): void {
        //Reset status variables
        this.endpointHasChanged = false;
        this.targetNameHasChanged = false;
        this.testOngoing = false;
        this.onGoing = false;

        //Reset data
        this.target = this.initEndpoint();
        this.initVal = this.initEndpoint();
        this.formValues = null;
    }

    //Forcely refresh the view
    forceRefreshView(duration: number): void {
        //Reset timer
        if (this.timerHandler) {
            clearInterval(this.timerHandler);
        }
        this.timerHandler = setInterval(() => this.ref.markForCheck(), 100);
        setTimeout(() => {
            if (this.timerHandler) {
                clearInterval(this.timerHandler);
                this.timerHandler = null;
            }
        }, duration);
    }

    openCreateEditTarget(editable: boolean, targetId?: number | string) {
        this.editable = editable;
        //reset
        this.reset();
        if (targetId) {
            this.actionType = ActionType.EDIT;
            this.translateService.get('DESTINATION.TITLE_EDIT').subscribe(res => this.modalTitle = res);
            toPromise<Endpoint>(this.endpointService
                .getEndpoint(targetId))
                .then(
                target => {
                    this.target = target;
                    //Keep data cache
                    this.initVal = clone(target);
                    this.initVal.password = FAKE_PASSWORD;
                    this.target.password = FAKE_PASSWORD;

                    //Open the modal now
                    this.open();
                    this.forceRefreshView(1000);
                })
                .catch(error => this.errorHandler.error(error));
        } else {
            this.actionType = ActionType.ADD_NEW;
            this.translateService.get('DESTINATION.TITLE_ADD').subscribe(res => this.modalTitle = res);
            //Directly open the modal
            this.open();
        }
    }

    testConnection() {
        let payload: Endpoint = this.initEndpoint();

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
                this.inlineAlert.showInlineSuccess({ message: "DESTINATION.TEST_CONNECTION_SUCCESS" });
                this.forceRefreshView(1000);
                this.testOngoing = false;
            }).catch(
            error => {
                this.inlineAlert.showInlineError('DESTINATION.TEST_CONNECTION_FAILURE');
                this.forceRefreshView(1000);
                this.testOngoing = false;
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
        if (this.onGoing) {
            return;//Avoid duplicated submitting
        }

        this.onGoing = true;
        toPromise<number>(this.endpointService
            .createEndpoint(this.target))
            .then(response => {
                this.translateService.get('DESTINATION.CREATED_SUCCESS')
                    .subscribe(res => this.errorHandler.info(res));
                this.reload.emit(true);
                this.onGoing = false;
                this.close();
            }).catch(error => {
                let errorMessageKey = this.handleErrorMessageKey(error.status);
                this.translateService
                    .get(errorMessageKey)
                    .subscribe(res => {
                        this.inlineAlert.showInlineError(res);
                        this.onGoing = false;
                    });
            }
            );
    }

    updateEndpoint() {
        if (this.onGoing) {
            return;//Avoid duplicated submitting
        }
        if (!(this.targetNameHasChanged || this.endpointHasChanged)) {
            return;//Avoid invalid submitting
        }
        let payload: Endpoint = this.initEndpoint();
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

        this.onGoing = true;
        toPromise<number>(this.endpointService
            .updateEndpoint(this.target.id, payload))
            .then(
            response => {
                this.translateService.get('DESTINATION.UPDATED_SUCCESS')
                    .subscribe(res => this.errorHandler.info(res));
                this.reload.emit(true);
                this.close();
                this.onGoing = false;
            })
            .catch(
            error => {
                let errorMessageKey = this.handleErrorMessageKey(error.status);
                this.translateService
                    .get(errorMessageKey)
                    .subscribe(res => {
                        this.inlineAlert.showInlineError(res);
                    });
                this.onGoing = false;
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
            this.close();
            if (this.targetForm) {
                this.targetForm.reset();
            }
        }
    }

    confirmCancel(confirmed: boolean) {
        this.inlineAlert.close();
        this.close();
    }

    ngAfterViewChecked(): void {
        if (this.targetForm != this.currentForm) {
            this.targetForm = this.currentForm;
            if (this.targetForm) {
                this.valueChangesSub = this.targetForm.valueChanges.subscribe((data: { [key: string]: string } | any) => {
                    if (data) {
                        //To avoid invalid change publish events
                        let keyNumber: number = 0;
                        for (let key in data) {
                            //Empty string "" is accepted
                            if (data[key] !== null) {
                                keyNumber++;
                            }
                        }
                        if (keyNumber !== 4) {
                            return;
                        }

                        if (!compareValue(this.formValues, data)) {
                            this.formValues = data;
                            this.inlineAlert.close();
                        }
                    }
                });
            }
        }
    }

}