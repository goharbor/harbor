import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import {
    Component,
    EventEmitter,
    Input,
    OnDestroy,
    OnInit,
    Output,
    ViewChild,
} from '@angular/core';
import { NgForm } from '@angular/forms';
import { TranslateService } from '@ngx-translate/core';
import { PreheatService } from '../../../../../../ng-swagger-gen/services/preheat.service';
import { Instance } from '../../../../../../ng-swagger-gen/models/instance';
import { AuthMode, FrontInstance } from '../distribution-interface';
import { clone } from '../../../../shared/units/utils';
import { ClrLoadingState } from '@clr/angular';
import { Metadata } from '../../../../../../ng-swagger-gen/models/metadata';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../shared/components/operation/operate';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import {
    debounceTime,
    distinctUntilChanged,
    filter,
    finalize,
    switchMap,
} from 'rxjs/operators';
import { Subject, Subscription } from 'rxjs';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { errorHandler } from '../../../../shared/units/shared.utils';

const DEFAULT_PROVIDER: string = 'dragonfly';

@Component({
    selector: 'dist-setup-modal',
    templateUrl: './distribution-setup-modal.component.html',
    styleUrls: ['./distribution-setup-modal.component.scss'],
})
export class DistributionSetupModalComponent implements OnInit, OnDestroy {
    @Input()
    providers: Metadata[] = [];
    model: Instance;
    originModelForEdit: Instance;
    opened: boolean = false;
    editingMode: boolean = false;
    authData: { [key: string]: any } = {};
    @ViewChild('instanceForm', { static: true }) instanceForm: NgForm;
    @ViewChild(InlineAlertComponent) inlineAlert: InlineAlertComponent;
    saveBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;

    @Output()
    refresh: EventEmitter<any> = new EventEmitter<any>();
    checkBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    onTesting: boolean = false;
    private _nameSubject: Subject<string> = new Subject<string>();
    private _endpointSubject: Subject<string> = new Subject<string>();
    private _nameSubscription: Subscription;
    private _endpointSubscription: Subscription;
    isNameExisting: boolean = false;
    isEndpointExisting: boolean = false;
    checkNameOnGoing: boolean = false;
    checkEndpointOngoing: boolean = false;
    constructor(
        private distributionService: PreheatService,
        private msgHandler: MessageHandlerService,
        private translate: TranslateService,
        private operationService: OperationService
    ) {}

    ngOnInit() {
        this.subscribeName();
        this.subscribeEndpoint();
        this.reset();
    }
    ngOnDestroy() {
        if (this._nameSubscription) {
            this._nameSubscription.unsubscribe();
            this._nameSubscription = null;
        }
        if (this._endpointSubscription) {
            this._endpointSubscription.unsubscribe();
            this._endpointSubscription = null;
        }
    }
    subscribeName() {
        if (!this._nameSubscription) {
            this._nameSubscription = this._nameSubject
                .pipe(
                    debounceTime(500),
                    distinctUntilChanged(),
                    filter(name => {
                        if (
                            this.editingMode &&
                            this.originModelForEdit &&
                            this.originModelForEdit.name === name
                        ) {
                            return false;
                        }
                        return name.length > 0;
                    }),
                    switchMap(name => {
                        this.isNameExisting = false;
                        this.checkNameOnGoing = true;
                        return this.distributionService
                            .ListInstances({
                                q: encodeURIComponent(`name=${name}`),
                            })
                            .pipe(
                                finalize(() => (this.checkNameOnGoing = false))
                            );
                    })
                )
                .subscribe(res => {
                    if (res && res.length > 0) {
                        this.isNameExisting = true;
                    }
                });
        }
    }
    subscribeEndpoint() {
        if (!this._endpointSubscription) {
            this._endpointSubscription = this._endpointSubject
                .pipe(
                    debounceTime(500),
                    distinctUntilChanged(),
                    filter(endpoint => {
                        if (
                            this.editingMode &&
                            this.originModelForEdit &&
                            this.originModelForEdit.endpoint === endpoint
                        ) {
                            return false;
                        }
                        return this.instanceForm.control.get('endpoint').valid;
                    }),
                    switchMap(endpoint => {
                        this.isEndpointExisting = false;
                        this.checkEndpointOngoing = true;
                        return this.distributionService
                            .ListInstances({
                                q: encodeURIComponent(`endpoint=${endpoint}`),
                            })
                            .pipe(
                                finalize(
                                    () => (this.checkEndpointOngoing = false)
                                )
                            );
                    })
                )
                .subscribe(res => {
                    if (res && res.length > 0) {
                        this.isEndpointExisting = true;
                    }
                });
        }
    }
    inputName() {
        this._nameSubject.next(this.model.name);
    }
    inputEndpoint() {
        this._endpointSubject.next(this.model.endpoint);
    }
    public get isValid(): boolean {
        return this.instanceForm && this.instanceForm.valid;
    }

    get title(): string {
        return this.editingMode
            ? 'DISTRIBUTION.EDIT_INSTANCE'
            : 'DISTRIBUTION.SETUP_NEW_INSTANCE';
    }

    authModeChange() {
        if (
            this.editingMode &&
            this.model.auth_mode === this.originModelForEdit.auth_mode
        ) {
            this.authData = clone(this.originModelForEdit.auth_info);
        } else {
            switch (this.model.auth_mode) {
                case AuthMode.BASIC:
                    this.authData = {
                        password: '',
                        username: '',
                    };
                    break;
                case AuthMode.OAUTH:
                    this.authData = {
                        token: '',
                    };
                    break;
                default:
                    this.authData = null;
                    break;
            }
        }
    }

    _open() {
        this.inlineAlert.close();
        this.opened = true;
    }

    _close() {
        this.opened = false;
        this.reset();
    }

    reset() {
        this.model = {
            name: '',
            endpoint: '',
            enabled: true,
            vendor: '',
            auth_mode: AuthMode.NONE,
            auth_info: this.authData,
        };
        this.instanceForm.reset({
            enabled: true,
        });
    }

    cancel() {
        this._close();
    }

    submit() {
        if (this.editingMode) {
            const operMessageForEdit = new OperateInfo();
            operMessageForEdit.name = 'DISTRIBUTION.UPDATE_INSTANCE';
            operMessageForEdit.data.id = this.model.id;
            operMessageForEdit.state = OperationState.progressing;
            operMessageForEdit.data.name = this.model.name;
            this.operationService.publishInfo(operMessageForEdit);
            this.saveBtnState = ClrLoadingState.LOADING;
            const instance: Instance = clone(this.originModelForEdit);
            instance.vendor = this.model.vendor;
            instance.name = this.model.name;
            instance.endpoint = this.model.endpoint;
            instance.insecure = this.model.insecure;
            instance.enabled = this.model.enabled;
            instance.auth_mode = this.model.auth_mode;
            instance.description = this.model.description;
            if (instance.auth_mode !== AuthMode.NONE) {
                instance.auth_info = this.authData;
            } else {
                delete instance.auth_info;
            }
            this.distributionService
                .UpdateInstance({
                    preheatInstanceName: this.model.name,
                    instance: this.handleInstance(instance),
                })
                .subscribe(
                    response => {
                        this.translate
                            .get('DISTRIBUTION.UPDATE_SUCCESS')
                            .subscribe(msg => {
                                operateChanges(
                                    operMessageForEdit,
                                    OperationState.success
                                );
                                this.msgHandler.info(msg);
                            });
                        this.saveBtnState = ClrLoadingState.SUCCESS;
                        this._close();
                        this.refresh.emit();
                    },
                    err => {
                        const message = errorHandler(err);
                        this.translate
                            .get('DISTRIBUTION.UPDATE_FAILED')
                            .subscribe(msg => {
                                this.translate
                                    .get(message)
                                    .subscribe(errMsg => {
                                        operateChanges(
                                            operMessageForEdit,
                                            OperationState.failure,
                                            msg
                                        );
                                        this.inlineAlert.showInlineError(
                                            msg + ': ' + errMsg
                                        );
                                        this.saveBtnState =
                                            ClrLoadingState.ERROR;
                                    });
                            });
                        this.msgHandler.handleErrorPopupUnauthorized(err);
                    }
                );
        } else {
            const operMessage = new OperateInfo();
            operMessage.name = 'DISTRIBUTION.CREATE_INSTANCE';
            operMessage.state = OperationState.progressing;
            operMessage.data.name = this.model.name;
            this.operationService.publishInfo(operMessage);
            this.saveBtnState = ClrLoadingState.LOADING;
            if (this.model.auth_mode !== AuthMode.NONE) {
                this.model.auth_info = this.authData;
            } else {
                delete this.model.auth_info;
            }
            // set insure property to true or false
            this.model.insecure = !!this.model.insecure;
            this.distributionService
                .CreateInstance({ instance: this.model })
                .subscribe(
                    response => {
                        this.translate
                            .get('DISTRIBUTION.CREATE_SUCCESS')
                            .subscribe(msg => {
                                operateChanges(
                                    operMessage,
                                    OperationState.success
                                );
                                this.msgHandler.info(msg);
                            });
                        this.saveBtnState = ClrLoadingState.SUCCESS;
                        this._close();
                        this.refresh.emit();
                    },
                    err => {
                        const message = errorHandler(err);
                        this.translate
                            .get('DISTRIBUTION.CREATE_FAILED')
                            .subscribe(msg => {
                                this.translate
                                    .get(message)
                                    .subscribe(errMsg => {
                                        operateChanges(
                                            operMessage,
                                            OperationState.failure,
                                            msg
                                        );
                                        this.inlineAlert.showInlineError(
                                            msg + ': ' + errMsg
                                        );
                                        this.saveBtnState =
                                            ClrLoadingState.ERROR;
                                    });
                            });
                        this.msgHandler.handleErrorPopupUnauthorized(err);
                    }
                );
        }
    }

    openSetupModal(editingMode: boolean, data?: Instance): void {
        this.editingMode = editingMode;
        this._open();
        if (editingMode) {
            this.model = clone(data);
            this.originModelForEdit = clone(data);
            // set insure property to true or false
            this.originModelForEdit.insecure = !!data.insecure;
            this.model.insecure = !!data.insecure;
            this.authData = this.model.auth_info || {};
        } else {
            this.reset();
            if (this.providers && this.providers.length) {
                this.providers.forEach(item => {
                    if (item.id === DEFAULT_PROVIDER) {
                        this.model.vendor = item.id;
                    }
                });
            }
        }
    }

    hasChangesForEdit(): boolean {
        if (this.editingMode) {
            if (this.model.vendor !== this.originModelForEdit.vendor) {
                return true;
            }
            if (this.model.name !== this.originModelForEdit.name) {
                return true;
            }
            if (
                this.model.description !== this.originModelForEdit.description
            ) {
                return true;
            }
            if (this.model.endpoint !== this.originModelForEdit.endpoint) {
                return true;
            }
            // eslint-disable-next-line eqeqeq
            if (this.model.enabled != this.originModelForEdit.enabled) {
                return true;
            }
            // eslint-disable-next-line eqeqeq
            if (this.model.insecure != this.originModelForEdit.insecure) {
                return true;
            }
            if (this.model.auth_mode !== this.originModelForEdit.auth_mode) {
                return true;
            } else {
                if (this.model.auth_mode === AuthMode.BASIC) {
                    if (
                        this.originModelForEdit.auth_info['username'] !==
                        this.authData['username']
                    ) {
                        return true;
                    }
                    if (
                        this.originModelForEdit.auth_info['password'] !==
                        this.authData['password']
                    ) {
                        return true;
                    }
                }
                if (this.model.auth_mode === AuthMode.OAUTH) {
                    if (
                        this.originModelForEdit.auth_info['token'] !==
                        this.authData['token']
                    ) {
                        return true;
                    }
                }
                return false;
            }
        }
        return true;
    }

    onTestEndpoint() {
        this.onTesting = true;
        this.checkBtnState = ClrLoadingState.LOADING;
        const instance: Instance = clone(this.model);
        instance.id = 0;
        this.distributionService
            .PingInstances({
                instance: this.handleInstance(instance),
            })
            .pipe(finalize(() => (this.onTesting = false)))
            .subscribe(
                res => {
                    this.checkBtnState = ClrLoadingState.SUCCESS;
                    this.inlineAlert.showInlineSuccess({
                        message: 'SCANNER.TEST_PASS',
                    });
                },
                error => {
                    this.inlineAlert.showInlineError(
                        'P2P_PROVIDER.TEST_FAILED'
                    );
                    this.checkBtnState = ClrLoadingState.ERROR;
                }
            );
    }
    handleInstance(instance: FrontInstance): Instance {
        if (instance) {
            const copyOne: FrontInstance = clone(instance);
            delete copyOne.hasCheckHealth;
            delete copyOne.pingStatus;
            return copyOne;
        }
        return instance;
    }
}
