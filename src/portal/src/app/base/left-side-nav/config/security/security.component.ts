import {
    Component,
    ElementRef,
    LOCALE_ID,
    OnDestroy,
    OnInit,
    ViewChild,
} from '@angular/core';
import { clone, compareValue } from '../../../../shared/units/utils';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import {
    ConfirmationState,
    ConfirmationTargets,
    DEFAULT_LANG_LOCALSTORAGE_KEY,
} from '../../../../shared/entities/shared.const';
import {
    SystemCVEAllowlist,
    SystemInfoService,
} from '../../../../shared/services';
import { Subscription } from 'rxjs';
import { finalize } from 'rxjs/operators';
import { ConfirmationDialogService } from '../../../global-confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../../../global-confirmation-dialog/confirmation-message';

const ONE_THOUSAND: number = 1000;
const CVE_DETAIL_PRE_URL = `https://nvd.nist.gov/vuln/detail/`;
const TARGET_BLANK = '_blank';

@Component({
    selector: 'app-security',
    templateUrl: './security.component.html',
    styleUrls: ['./security.component.scss'],
    providers: [
        {
            provide: LOCALE_ID,
            useValue: localStorage.getItem(DEFAULT_LANG_LOCALSTORAGE_KEY),
        },
    ],
})
export class SecurityComponent implements OnInit, OnDestroy {
    onGoing = false;
    systemAllowlist: SystemCVEAllowlist;
    systemAllowlistOrigin: SystemCVEAllowlist;
    cveIds: string;
    showAddModal: boolean = false;
    @ViewChild('dateInput') dateInput: ElementRef;
    private _confirmSub: Subscription;
    constructor(
        private errorHandler: ErrorHandler,
        private systemInfoService: SystemInfoService,
        private confirmService: ConfirmationDialogService
    ) {}
    ngOnInit() {
        this.getSystemAllowlist();
        this.subscribeConfirmation();
    }
    ngOnDestroy() {
        if (this._confirmSub) {
            this._confirmSub.unsubscribe();
            this._confirmSub = null;
        }
    }
    save(): void {
        if (!compareValue(this.systemAllowlistOrigin, this.systemAllowlist)) {
            this.onGoing = true;
            this.systemInfoService
                .updateSystemAllowlist(this.systemAllowlist)
                .pipe(finalize(() => (this.onGoing = false)))
                .subscribe({
                    next: res => {
                        this.systemAllowlistOrigin = clone(
                            this.systemAllowlist
                        );
                        this.errorHandler.info('CONFIG.SAVE_SUCCESS');
                    },
                    error: err => {
                        this.errorHandler.error(err);
                    },
                });
        } else {
            // Inprop situation, should not come here
            console.error('Save abort because nothing changed');
        }
    }

    get inProgress(): boolean {
        return this.onGoing;
    }

    cancel(): void {
        if (!compareValue(this.systemAllowlistOrigin, this.systemAllowlist)) {
            const msg = new ConfirmationMessage(
                'CONFIG.CONFIRM_TITLE',
                'CONFIG.CONFIRM_SUMMARY',
                '',
                null,
                ConfirmationTargets.CONFIG
            );
            this.confirmService.openComfirmDialog(msg);
        } else {
            // Invalid situation, should not come here
            console.error('Nothing changed');
        }
    }

    subscribeConfirmation() {
        this._confirmSub = this.confirmService.confirmationConfirm$.subscribe(
            confirmation => {
                if (
                    confirmation &&
                    confirmation.state === ConfirmationState.CONFIRMED
                ) {
                    if (
                        !compareValue(
                            this.systemAllowlistOrigin,
                            this.systemAllowlist
                        )
                    ) {
                        this.systemAllowlist = clone(
                            this.systemAllowlistOrigin
                        );
                    }
                }
            }
        );
    }

    getSystemAllowlist() {
        this.onGoing = true;
        this.systemInfoService.getSystemAllowlist().subscribe(
            systemAllowlist => {
                this.onGoing = false;
                if (!systemAllowlist) {
                    systemAllowlist = {};
                }
                if (!systemAllowlist.items) {
                    systemAllowlist.items = [];
                }
                if (!systemAllowlist.expires_at) {
                    systemAllowlist.expires_at = null;
                }
                this.systemAllowlist = systemAllowlist;
                this.systemAllowlistOrigin = clone(systemAllowlist);
            },
            error => {
                this.onGoing = false;
                console.error(
                    'An error occurred during getting systemAllowlist'
                );
            }
        );
    }

    deleteItem(index: number) {
        this.systemAllowlist.items.splice(index, 1);
    }

    addToSystemAllowlist() {
        // remove duplication and add to systemAllowlist
        let map = {};
        this.systemAllowlist.items.forEach(item => {
            map[item.cve_id] = true;
        });
        this.cveIds.split(/[\n,]+/).forEach(id => {
            let cveObj: any = {};
            cveObj.cve_id = id.trim();
            if (!map[cveObj.cve_id]) {
                map[cveObj.cve_id] = true;
                this.systemAllowlist.items.push(cveObj);
            }
        });
        // clear modal and close modal
        this.cveIds = null;
        this.showAddModal = false;
    }

    get hasAllowlistChanged(): boolean {
        return !compareValue(this.systemAllowlistOrigin, this.systemAllowlist);
    }

    isDisabled(): boolean {
        let str = this.cveIds;
        return !(str && str.trim());
    }

    get expiresDate() {
        if (this.systemAllowlist && this.systemAllowlist.expires_at) {
            return new Date(this.systemAllowlist.expires_at * ONE_THOUSAND);
        }
        return null;
    }

    set expiresDate(date) {
        if (this.systemAllowlist && date) {
            this.systemAllowlist.expires_at = Math.floor(
                date.getTime() / ONE_THOUSAND
            );
        }
    }

    get neverExpires(): boolean {
        return !(this.systemAllowlist && this.systemAllowlist.expires_at);
    }

    set neverExpires(flag) {
        if (flag) {
            this.systemAllowlist.expires_at = null;
            this.systemInfoService.resetDateInput(this.dateInput);
        } else {
            this.systemAllowlist.expires_at = Math.floor(
                new Date().getTime() / ONE_THOUSAND
            );
        }
    }

    get hasExpired(): boolean {
        if (
            this.systemAllowlistOrigin &&
            this.systemAllowlistOrigin.expires_at
        ) {
            return (
                new Date().getTime() >
                this.systemAllowlistOrigin.expires_at * ONE_THOUSAND
            );
        }
        return false;
    }

    goToDetail(cveId) {
        window.open(CVE_DETAIL_PRE_URL + `${cveId}`, TARGET_BLANK);
    }
}
