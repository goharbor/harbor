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

import { AfterViewChecked, Component, ErrorHandler, OnInit, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';
import { NavigationExtras, Router } from '@angular/router';
import { SessionUser } from '../../shared/entities/session-user';
import { SessionService } from '../../shared/services/session.service';
import { MessageHandlerService } from '../../shared/services/message-handler.service';
import { SearchTriggerService } from '../../shared/components/global-search/search-trigger.service';
import { ResetSecret } from './account';
import { CopyInputComponent } from '../../shared/components/push-image/copy-input.component';
import {
    CommonRoutes,
    ConfirmationButtons,
    ConfirmationTargets,
    ConfirmationState,
} from '../../shared/entities/shared.const';
import { ConfirmationDialogComponent } from '../../shared/components/confirmation-dialog';
import { InlineAlertComponent } from '../../shared/components/inline-alert/inline-alert.component';
import { ConfirmationMessage } from '../global-confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../global-confirmation-dialog/confirmation-state-message';
import { UserService } from 'ng-swagger-gen/services/user.service';
import { AppConfigService } from '../../services/app-config.service';

@Component({
    selector: 'account-settings-modal',
    templateUrl: 'account-settings-modal.component.html',
    styleUrls: ['./account-settings-modal.component.scss', '../../common.scss'],
    standalone: false,
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

    // PAT Management
    pats: any[] = [];
    selectedPATs: any[] = [];
    patLoading: boolean = false;
    showCreatePATModal: boolean = false;
    newPATForm = { name: '', expiresInDays: 0, description: '' };
    createdPATSecret: string = '';

    constructor(
        private session: SessionService,
        private msgHandler: MessageHandlerService,
        private router: Router,
        private searchTrigger: SearchTriggerService,
        private userService: UserService,
        private appConfigService: AppConfigService
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
                    if (this.account.username === this.originAdminName && this.inlineAlert) {
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

                        // Mail changed, if self-registration disabled, only system admin can check mail-existing status
                        if (
                            this.session.getCurrentUser()?.has_admin_role ||
                            this.appConfigService.getConfig()?.self_registration
                        ) {
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
                    } else if (this.inlineAlert) {
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
        if (this.inlineAlert) {
            this.inlineAlert.close();
        }

        // Clear check history
        this.mailAlreadyChecked = {};

        // Reset validation status
        this.validationStateMap = {
            account_settings_email: true,
            account_settings_full_name: true,
        };
        this.showGenerateCli = false;

        // Open modal — clrDgRefresh event will load PATs automatically
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
                } else if (this.inlineAlert) {
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
            if (this.inlineAlert) {
                this.inlineAlert.showInlineWarning({
                    message: 'PROFILE.RENAME_CONFIRM_INFO',
                });
            }
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
                    } else if (this.inlineAlert) {
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
        if (this.inlineAlert) {
            this.inlineAlert.close();
        }
        this.opened = false;
    }

    onSuccess(event) {
        if (this.inlineAlert) {
            this.inlineAlert.showInlineSuccess({ message: 'PROFILE.COPY_SUCCESS' });
        }
    }

    onError(event) {
        if (this.inlineAlert) {
            this.inlineAlert.showInlineError({ message: 'PROFILE.COPY_ERROR' });
        }
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

    confirmGenerate(): void {
        const generatedSecret = this.generateRandomSecret();
        this.resetCliSecret(generatedSecret);
    }

    private generateRandomSecret(): string {
        const chars =
            'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
        let result = '';
        for (let i = 0; i < 16; i++) {
            result += chars.charAt(Math.floor(Math.random() * chars.length));
        }
        // Ensure requirements: 8-128 chars, at least 1 uppercase, 1 lowercase, 1 digit
        // Replace some random positions with required characters
        const resultArray = result.split('');
        resultArray[0] = 'A'; // uppercase
        resultArray[1] = 'a'; // lowercase
        resultArray[2] = '1'; // digit
        return resultArray.join('');
    }

    resetCliSecret(secret) {
        this.userService
            .setCliSecret({
                userId: this.account.user_id,
                secret: secret
                    ? {
                          secret: secret,
                      }
                    : {},
            })
            .subscribe({
                next: res => {
                    if (secret) {
                        this.account.oidc_user_meta.secret = secret;
                    } else {
                        this.userService.getCurrentUserInfo().subscribe(res => {
                            this.account.oidc_user_meta.secret =
                                res?.oidc_user_meta?.secret;
                        });
                    }
                    this.closeReset();
                    if (this.inlineAlert) {
                        this.inlineAlert.showInlineSuccess({
                            message: 'PROFILE.GENERATE_SUCCESS',
                        });
                    }
                },
                error: err => {
                    if (this.resetSecretInlineAlert) {
                        this.resetSecretInlineAlert.showInlineError({
                            message: 'PROFILE.GENERATE_ERROR',
                        });
                    }
                },
            });
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
        if (this.resetSecretInlineAlert) {
            this.resetSecretInlineAlert.close();
        }
    }

    // PAT Management Methods
    onPATsRefresh(state: any): void {
        this.loadPATs();
    }

    openCreatePATModal() {
        this.showCreatePATModal = true;
        this.newPATForm = { name: '', expiresInDays: 0, description: '' };
        this.createdPATSecret = '';
    }

    closeCreatePATModal() {
        this.showCreatePATModal = false;
        this.loadPATs();
    }

    loadPATs() {
        if (!this.account) {
            return;
        }
        this.patLoading = true;
        this.userService
            .ListPersonalAccessTokens({ userId: this.account.user_id })
            .subscribe({
                next: (res: any) => {
                    try {
                        this.pats = Array.isArray(res) ? res : [];
                        this.pats.forEach(pat => {
                            pat.expired =
                                pat.expires_at > 0 &&
                                pat.expires_at <= Date.now() / 1000;
                        });
                        console.log('[PAT] Data assigned, count:', this.pats.length);
                    } catch (e) {
                        console.error('[PAT] Error during data assignment:', e);
                        throw e;
                    }
                    this.patLoading = false;
                },
                error: (err: any) => {
                    console.error('[PAT] Error loading PATs:', err);
                    this.pats = [];
                    this.patLoading = false;
                    this.msgHandler.handleError(err);
                },
            });
    }

    copyPATSecret() {
        if (this.createdPATSecret) {
            navigator.clipboard.writeText(this.createdPATSecret);
        }
    }

    createPAT() {
        if (!this.newPATForm.name || !this.account) {
            return;
        }
        this.userService
            .CreatePersonalAccessToken({
                userId: this.account.user_id,
                request: {
                    name: this.newPATForm.name,
                    description: this.newPATForm.description,
                    expires_in_days: this.newPATForm.expiresInDays,
                },
            })
            .subscribe({
                next: (res: any) => {
                    this.createdPATSecret = res.secret;
                    this.msgHandler.showSuccess('PROFILE.PAT_CREATE_SUCCESS');
                    this.loadPATs();
                },
                error: (err: any) => {
                    const status = err?.status || err?.error?.status;
                    if (status === 409) {
                        this.msgHandler.showError(
                            'PROFILE.PAT_NAME_CONFLICT',
                            null
                        );
                    } else {
                        this.msgHandler.handleError(err);
                    }
                },
            });
    }

    refreshPATSecret(patId: number) {
        if (!this.account) {
            return;
        }
        this.userService
            .RefreshPersonalAccessTokenSecret({
                userId: this.account.user_id,
                tokenId: patId,
                request: {},
            })
            .subscribe({
                next: (res: any) => {
                    this.createdPATSecret = res.secret;
                    this.showCreatePATModal = true;
                    this.msgHandler.showSuccess('PROFILE.PAT_REFRESHED');
                    this.loadPATs();
                },
                error: (err: any) => {
                    this.msgHandler.handleError(err);
                },
            });
    }

    togglePATDisabled(pat: any) {
        if (!this.account) {
            return;
        }
        this.userService
            .UpdatePersonalAccessToken({
                userId: this.account.user_id,
                tokenId: pat.id,
                request: {
                    disabled: !pat.disabled,
                },
            })
            .subscribe({
                next: () => {
                    this.msgHandler.showSuccess('PROFILE.PAT_UPDATED');
                    this.loadPATs();
                },
                error: (err: any) => {
                    this.msgHandler.handleError(err);
                },
            });
    }

    deletePAT(patId: number) {
        if (!this.account) {
            return;
        }
        const deletePatMessage: ConfirmationMessage = new ConfirmationMessage(
            'PROFILE.DELETE_PAT_TITLE',
            'PROFILE.DELETE_PAT_CONFIRM',
            'BUTTON.DELETE',
            'BUTTON.CANCEL',
            ConfirmationTargets.USER_PAT
        );
        deletePatMessage.data = patId;
        this.confirmationDialogComponent.open(deletePatMessage);
    }

    confirmAction(message: ConfirmationAcknowledgement) {
        if (!message || message.state !== ConfirmationState.CONFIRMED) {
            return;
        }
        if (message.source === ConfirmationTargets.USER_PAT) {
            this.confirmDeletePAT(message.data);
        } else {
            this.confirmGenerate();
        }
    }

    confirmDeletePAT(patId: number) {
        if (!this.account || !patId) {
            return;
        }
        this.userService
            .DeletePersonalAccessToken({
                userId: this.account.user_id,
                tokenId: patId,
            })
            .subscribe({
                next: () => {
                    this.msgHandler.showSuccess('PROFILE.PAT_DELETED');
                    this.loadPATs();
                    this.selectedPATs = [];
                },
                error: (err: any) => {
                    this.msgHandler.handleError(err);
                },
            });
    }
}
