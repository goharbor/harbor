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
    OnInit,
    ViewChild,
    OnDestroy,
    ElementRef,
    ChangeDetectorRef,
} from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { forkJoin, Observable, Subscription } from 'rxjs';
import { AppConfigService } from '../../services/app-config.service';
import { ModalEvent } from '../modal-event';
import { modalEvents } from '../modal-events.const';
import { PasswordSettingComponent } from '../password-setting/password-setting.component';
import { NavigatorComponent } from '../../shared/components/navigator/navigator.component';
import { SessionService } from '../../shared/services/session.service';
import { AboutDialogComponent } from '../../shared/components/about-dialog/about-dialog.component';
import { SearchTriggerService } from '../../shared/components/global-search/search-trigger.service';
import { CommonRoutes } from '../../shared/entities/shared.const';
import { THEME_ARRAY, ThemeInterface } from '../../services/theme';
import { clone, DEFAULT_PAGE_SIZE } from '../../shared/units/utils';
import { ThemeService } from '../../services/theme.service';
import { AccountSettingsModalComponent } from '../account-settings/account-settings-modal.component';
import {
    EventService,
    HarborEvent,
} from '../../services/event-service/event.service';
import { SCANNERS_DOC } from '../left-side-nav/interrogation-services/scanner/scanner';
import { ScannerService } from '../../../../ng-swagger-gen/services/scanner.service';
import { Project } from '../../../../ng-swagger-gen/models/project';

const HAS_SHOWED_SCANNER_INFO: string = 'hasShowScannerInfo';
const YES: string = 'yes';
const HAS_STYLE_MODE: string = 'styleModeLocal';

@Component({
    selector: 'harbor-shell',
    templateUrl: 'harbor-shell.component.html',
    styleUrls: ['harbor-shell.component.scss'],
})
export class HarborShellComponent implements OnInit, OnDestroy {
    @ViewChild(AccountSettingsModalComponent)
    accountSettingsModal: AccountSettingsModalComponent;

    @ViewChild(PasswordSettingComponent)
    pwdSetting: PasswordSettingComponent;

    @ViewChild(NavigatorComponent)
    navigator: NavigatorComponent;

    @ViewChild(AboutDialogComponent)
    aboutDialog: AboutDialogComponent;

    // To indicator whwther or not the search results page is displayed
    // We need to use this property to do some overriding work
    isSearchResultsOpened: boolean = false;

    searchSub: Subscription;
    searchCloseSub: Subscription;
    isLdapMode: boolean;
    isOidcMode: boolean;
    isHttpAuthMode: boolean;
    showScannerInfo: boolean = false;
    scannerDocUrl: string = SCANNERS_DOC;
    themeArray: ThemeInterface[] = clone(THEME_ARRAY);

    styleMode = this.themeArray[0].showStyle;
    @ViewChild('scrollDiv') scrollDiv: ElementRef;
    scrollToPositionSub: Subscription;
    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private session: SessionService,
        private searchTrigger: SearchTriggerService,
        private appConfigService: AppConfigService,
        private scannerService: ScannerService,
        public theme: ThemeService,
        private event: EventService,
        private cd: ChangeDetectorRef
    ) {}

    ngOnInit() {
        if (!this.scrollToPositionSub) {
            this.scrollToPositionSub = this.event.subscribe(
                HarborEvent.SCROLL_TO_POSITION,
                scrollTop => {
                    if (this.scrollDiv && this.scrollDiv.nativeElement) {
                        this.cd.detectChanges();
                        this.scrollDiv.nativeElement.scrollTop = scrollTop;
                    }
                }
            );
        }
        if (this.appConfigService.isLdapMode()) {
            this.isLdapMode = true;
        } else if (this.appConfigService.isHttpAuthMode()) {
            this.isHttpAuthMode = true;
        } else if (this.appConfigService.isOidcMode()) {
            this.isOidcMode = true;
        }
        this.searchSub = this.searchTrigger.searchTriggerChan$.subscribe(
            searchEvt => {
                if (searchEvt && searchEvt.trim() !== '') {
                    this.isSearchResultsOpened = true;
                }
            }
        );

        this.searchCloseSub = this.searchTrigger.searchCloseChan$.subscribe(
            close => {
                this.isSearchResultsOpened = false;
            }
        );
        if (
            !(
                localStorage &&
                localStorage.getItem(HAS_SHOWED_SCANNER_INFO) === YES
            )
        ) {
            if (this.isSystemAdmin) {
                this.getDefaultScanner();
            }
        }
        // set local in app
        if (localStorage) {
            this.styleMode = localStorage.getItem(HAS_STYLE_MODE);
        }
    }
    publishScrollEvent() {
        if (this.scrollDiv && this.scrollDiv.nativeElement) {
            this.event.publish(HarborEvent.SCROLL, {
                url: this.router.url,
                scrollTop: this.scrollDiv.nativeElement.scrollTop,
            });
        }
    }
    closeInfo() {
        if (localStorage) {
            localStorage.setItem(HAS_SHOWED_SCANNER_INFO, YES);
        }
        this.showScannerInfo = false;
    }

    getDefaultScanner() {
        this.scannerService
            .listScannersResponse({
                pageSize: DEFAULT_PAGE_SIZE,
                page: 1,
            })
            .subscribe(res => {
                if (res.headers) {
                    const xHeader: string = res.headers.get('X-Total-Count');
                    const totalCount = parseInt(xHeader, 0);
                    let arr = res.body || [];
                    if (totalCount <= DEFAULT_PAGE_SIZE) {
                        // already gotten all scanners
                        if (arr && arr.length) {
                            this.showScannerInfo = arr.some(
                                scanner => scanner.is_default
                            );
                        }
                    } else {
                        // get all the scanners in specified times
                        const times: number = Math.ceil(
                            totalCount / DEFAULT_PAGE_SIZE
                        );
                        const observableList: Observable<Project[]>[] = [];
                        for (let i = 2; i <= times; i++) {
                            observableList.push(
                                this.scannerService.listScanners({
                                    page: i,
                                    pageSize: DEFAULT_PAGE_SIZE,
                                })
                            );
                        }
                        forkJoin(observableList).subscribe(response => {
                            if (response && response.length) {
                                response.forEach(item => {
                                    arr = arr.concat(item);
                                });
                                this.showScannerInfo = arr.some(
                                    scanner => scanner.is_default
                                );
                            }
                        });
                    }
                }
            });
    }
    ngOnDestroy(): void {
        if (this.searchSub) {
            this.searchSub.unsubscribe();
        }

        if (this.searchCloseSub) {
            this.searchCloseSub.unsubscribe();
        }
        if (this.scrollToPositionSub) {
            this.scrollToPositionSub.unsubscribe();
            this.scrollToPositionSub = null;
        }
    }

    public get shouldOverrideContent(): boolean {
        return this.router.routerState.snapshot.url
            .toString()
            .startsWith(CommonRoutes.EMBEDDED_SIGN_IN);
    }

    public get showSearch(): boolean {
        return this.isSearchResultsOpened;
    }

    public get isSystemAdmin(): boolean {
        let account = this.session.getCurrentUser();
        return account != null && account.has_admin_role;
    }

    public get isUserExisting(): boolean {
        let account = this.session.getCurrentUser();
        return account != null;
    }
    public get hasAdminRole(): boolean {
        return (
            this.session.getCurrentUser() &&
            this.session.getCurrentUser().has_admin_role
        );
    }
    public get withAdmiral(): boolean {
        return this.appConfigService.getConfig().with_admiral;
    }
    // Open modal dialog
    openModal(event: ModalEvent): void {
        switch (event.modalName) {
            case modalEvents.USER_PROFILE:
                this.accountSettingsModal.open();
                break;
            case modalEvents.CHANGE_PWD:
                this.pwdSetting.open();
                break;
            case modalEvents.ABOUT:
                this.aboutDialog.open();
                break;
            default:
                break;
        }
    }
    themeChanged(theme) {
        this.styleMode = theme.mode;
        this.theme.loadStyle(theme.toggleFileName);
        if (localStorage) {
            localStorage.setItem(HAS_STYLE_MODE, this.styleMode);
        }
    }
}
