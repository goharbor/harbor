// Copyright Project Harbor Authors
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

import { AfterViewChecked, Component, OnInit, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';
import { NavigationExtras, Router } from '@angular/router';
import { SessionUser } from '../../shared/entities/session-user';
import { SessionService } from '../../shared/services/session.service';
import { MessageHandlerService } from '../../shared/services/message-handler.service';
import { SearchTriggerService } from '../../shared/components/global-search/search-trigger.service';
import { AccountSettingsModalService } from './account-settings-modal-service.service';
import { randomWord } from '../../shared/units/shared.utils';
import { ResetSecret } from './account';
import { CopyInputComponent } from '../../shared/components/push-image/copy-input.component';
import {
    CommonRoutes,
    ConfirmationButtons,
    ConfirmationTargets,
} from '../../shared/entities/shared.const';
import { ConfirmationDialogComponent } from '../../shared/components/confirmation-dialog';
import { InlineAlertComponent } from '../../shared/components/inline-alert/inline-alert.component';
import { ConfirmationMessage } from '../global-confirmation-dialog/confirmation-message';

@Component({
    selector: 'account-settings-modal',
    templateUrl: 'account-settings-modal.component.html',
    styleUrls: ['./account-settings-modal.component.scss', '../../common.scss'],
})
export class AccountSettingsModalComponent implements OnInit, AfterViewChecked {
    opened = false;
    staticBackdrop = true;
    originalStaticData: SessionUser;
    account: SessionUser;
    error: any = null;
    emailTooltip = 'TOOLTIP.EMAIL';
    mailAlreadyChecked = {};
    isOnCalling = false;
    formValueChanged = false;
    checkOnGoing = false;
    RenameOnGoing = false;
    originAdminName = 'admin';
    newAdminName = 'admin@harbor.local';
    renameConfirmation = false;
    showSecretDetail = false;
    resetForms = new ResetSecret();
    showGenerateCli: boolean = false;
    @ViewChild('confirmationDialog')
    confirmationDialogComponent: ConfirmationDialogComponent;

    accountFormRef: NgForm;
    @ViewChild('accountSettingsFrom', { static: true }) accountForm: NgForm;
    @ViewChild('resetSecretFrom', { static: true }) resetSecretFrom: NgForm;
    @ViewChild('accountSettingInlineAlert') inlineAlert: InlineAlertComponent;
    @ViewChild('resetSecretInlineAlert')
    resetSecretInlineAlert: InlineAlertComponent;
    @ViewChild('copyInput') copyInput: CopyInputComponent;
    showInputSecret: boolean = false;
    showConfirmSecret: boolean = false;

    constructor(
        private session: SessionService,
        private msgHandler: MessageHandlerService,
        private router: Router,
        private searchTrigger: SearchTriggerService,
        private accountSettingsService: AccountSettingsModalService
    ) {}

    private validationStateMap: any = {
        account_settings_email: true,
        account_settings_full_name: true,
    };

    ngOnInit(): void {
        this.refreshAccount();
    }

    refreshAccount() {
        // Value copy
        this.account = Object.assign({}, this.session.getCurrentUser());
        this.originalStaticData = Object.assign(
            {},
            this.session.getCurrentUser()
        );
    }

    ngAfterViewChecked(): void {
        if (this.accountFormRef !== this.accountForm) {
            this.accountFormRef = this.accountForm;
            if (this.accountFormRef) {
                this.accountFormRef.valueChanges.subscribe(data => {
                    if (this.error) {
                        this.error = null;
                    }
                    this.formValueChanged = true;
                    if (this.account.username === this.originAdminName) {
                        this.inlineAlert.close();
                    }
                });
            }
        }
    }

    getValidationState(key: string): boolean {
        return this.validationStateMap[key];
    }

    handleValidation(key: string, flag: boolean): void {
        if (flag) {
            // Checking
            let cont = this.accountForm.controls[key];
            if (cont) {
                this.validationStateMap[key] = cont.valid;
                // Check email existing from backend
                if (cont.valid && key === 'account_settings_email') {
                    if (
                        this.formValueChanged &&
                        this.account.email !== this.originalStaticData.email
                    ) {
                        if (this.mailAlreadyChecked[this.account.email]) {
                            this.validationStateMap[key] =
                                !this.mailAlreadyChecked[this.account.email]
                                    .result;
                            if (!this.validationStateMap[key]) {
                                this.emailTooltip = 'TOOLTIP.EMAIL_EXISTING';
                            }
                            return;
                        }

                        // Mail changed
                        this.checkOnGoing = true;
                        this.session
                            .checkUserExisting('email', this.account.email)
                            .subscribe(
                                (res: boolean) => {
                                    this.checkOnGoing = false;
                                    this.validationStateMap[key] = !res;
                                    if (res) {
                                        this.emailTooltip =
                                            'TOOLTIP.EMAIL_EXISTING';
                                    }
                                    this.mailAlreadyChecked[
                                        this.account.email
                                    ] = {
                                        result: res,
                                    }; // Tag it checked
                                },
                                error => {
                                    this.checkOnGoing = false;
                                    this.validationStateMap[key] = false; // Not valid @ backend
                                }
                            );
                    }
                }
            }
        } else {
            // Reset
            this.validationStateMap[key] = true;
            this.emailTooltip = 'TOOLTIP.EMAIL';
        }
    }

    isUserDataChange(): boolean {
        if (!this.originalStaticData || !this.account) {
            return false;
        }
        for (let prop in this.originalStaticData) {
            if (this.originalStaticData[prop] !== this.account[prop]) {
                return true;
            }
        }
        return false;
    }

    public get isValid(): boolean {
        return (
            this.accountForm &&
            this.accountForm.valid &&
            this.error === null &&
            this.validationStateMap['account_settings_email']
        ); // backend check is valid as well
    }

    public get showProgress(): boolean {
        return this.isOnCalling;
    }

    public get checkProgress(): boolean {
        return this.checkOnGoing;
    }

    public get canRename(): boolean {
        return (
            this.account &&
            this.account.has_admin_role &&
            this.originalStaticData.username === 'admin' &&
            this.account.user_id === 1
        );
    }

    onRename(): void {
        this.account.username = this.newAdminName;
        this.RenameOnGoing = true;
    }

    confirmRename(): void {
        if (this.canRename) {
            this.session.updateAccountSettings(this.account).subscribe(
                () => {
                    this.session.renameAdmin(this.account).subscribe(
                        () => {
                            this.msgHandler.showSuccess(
                                'PROFILE.RENAME_SUCCESS'
                            );
                            this.opened = false;
                            this.logOut();
                        },
                        error => {
                            this.msgHandler.handleError(error);
                        }
                    );
                },
                error => {
                    this.isOnCalling = false;
                    this.error = error;
                    if (this.msgHandler.isAppLevel(error)) {
                        this.opened = false;
                        this.msgHandler.handleError(error);
                    } else {
                        this.inlineAlert.showInlineError(error);
                    }
                }
            );
        }
    }

    // Log out system
    logOut(): void {
        // Naviagte to the sign in router-guard
        // Appending 'signout' means destroy session cache
        let navigatorExtra: NavigationExtras = {
            queryParams: { signout: true },
        };
        this.router.navigate([CommonRoutes.EMBEDDED_SIGN_IN], navigatorExtra);
        // Confirm search result panel is close
        this.searchTrigger.closeSearch(true);
    }

    open() {
        // Keep the initial data for future diff
        this.originalStaticData = Object.assign(
            {},
            this.session.getCurrentUser()
        );
        this.account = Object.assign({}, this.session.getCurrentUser());
        this.formValueChanged = false;

        // Confirm inline alert is closed
        this.inlineAlert.close();

        // Clear check history
        this.mailAlreadyChecked = {};

        // Reset validation status
        this.validationStateMap = {
            account_settings_email: true,
            account_settings_full_name: true,
        };
        this.showGenerateCli = false;
        this.opened = true;
    }

    close() {
        if (this.formValueChanged) {
            if (!this.isUserDataChange()) {
                this.opened = false;
            } else {
                if (this.RenameOnGoing) {
                    this.RenameOnGoing = false;
                    this.opened = false;
                } else {
                    // Need user confirmation
                    this.inlineAlert.showInlineConfirmation({
                        message: 'ALERT.FORM_CHANGE_CONFIRMATION',
                    });
                }
            }
        } else {
            this.opened = false;
        }
    }

    submit() {
        if (!this.isValid || this.isOnCalling) {
            return;
        }

        // Double confirm session is valid
        let cUser = this.session.getCurrentUser();
        if (!cUser) {
            return;
        }

        if (this.RenameOnGoing && !this.renameConfirmation) {
            this.renameConfirmation = true;
            this.inlineAlert.showInlineWarning({
                message: 'PROFILE.RENAME_CONFIRM_INFO',
            });
            return;
        }

        this.isOnCalling = true;

        if (this.RenameOnGoing && this.renameConfirmation) {
            this.confirmRename();
        } else {
            this.session.updateAccountSettings(this.account).subscribe(
                () => {
                    this.isOnCalling = false;
                    this.opened = false;
                    this.msgHandler.showSuccess('PROFILE.SAVE_SUCCESS');
                    // get user info from back-end then refresh account
                    this.session.retrieveUser().subscribe(() => {
                        this.refreshAccount();
                    });
                },
                error => {
                    this.isOnCalling = false;
                    this.error = error;
                    if (this.msgHandler.isAppLevel(error)) {
                        this.opened = false;
                        this.msgHandler.handleError(error);
                    } else {
                        this.inlineAlert.showInlineError(error);
                    }
                }
            );
        }
    }

    confirmNo($event: any): void {
        if (this.RenameOnGoing) {
            this.RenameOnGoing = false;
        }
        if (this.renameConfirmation) {
            this.renameConfirmation = false;
        }
    }

    confirmYes($event: any): void {
        if (this.RenameOnGoing) {
            this.RenameOnGoing = false;
        }
        if (this.renameConfirmation) {
            this.renameConfirmation = false;
        }
        this.inlineAlert.close();
        this.opened = false;
    }

    onSuccess(event) {
        this.inlineAlert.showInlineSuccess({ message: 'PROFILE.COPY_SUCCESS' });
    }

    onError(event) {
        this.inlineAlert.showInlineError({ message: 'PROFILE.COPY_ERROR' });
    }

    generateCli(userId): void {
        let generateCliMessage = new ConfirmationMessage(
            'PROFILE.CONFIRM_TITLE_CLI_GENERATE',
            'PROFILE.CONFIRM_BODY_CLI_GENERATE',
            '',
            userId,
            ConfirmationTargets.TARGET,
            ConfirmationButtons.CONFIRM_CANCEL
        );
        this.confirmationDialogComponent.open(generateCliMessage);
    }

    showGenerateCliFn() {
        this.showGenerateCli = !this.showGenerateCli;
    }

    confirmGenerate(event): void {
        this.account.oidc_user_meta.secret = randomWord(9);
        this.resetCliSecret(this.account.oidc_user_meta.secret);
    }

    resetCliSecret(secret) {
        let userId = this.account.user_id;
        this.accountSettingsService
            .saveNewCli(userId, { secret: secret })
            .subscribe(
                cliSecret => {
                    this.account.oidc_user_meta.secret = secret;
                    this.closeReset();
                    this.inlineAlert.showInlineSuccess({
                        message: 'PROFILE.GENERATE_SUCCESS',
                    });
                },
                error => {
                    this.resetSecretInlineAlert.showInlineError({
                        message: 'PROFILE.GENERATE_ERROR',
                    });
                }
            );
    }

    disableChangeCliSecret() {
        return (
            this.resetSecretFrom.invalid ||
            this.resetSecretFrom.value.input_secret !==
                this.resetSecretFrom.value.confirm_secret
        );
    }

    closeReset() {
        this.showSecretDetail = false;
        this.showGenerateCliFn();
        this.resetSecretFrom.resetForm(new ResetSecret());
    }

    openSecretDetail() {
        this.showSecretDetail = true;
        this.resetSecretInlineAlert.close();
    }
}
