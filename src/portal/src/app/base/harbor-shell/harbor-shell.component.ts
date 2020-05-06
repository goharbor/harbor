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
import { Component, OnInit, ViewChild, OnDestroy } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { Subscription } from "rxjs";
import { AppConfigService } from '../../services/app-config.service';

import { ModalEvent } from '../modal-event';
import { modalEvents } from '../modal-events.const';

import { AccountSettingsModalComponent } from '../../account/account-settings/account-settings-modal.component';
import { PasswordSettingComponent } from '../../account/password-setting/password-setting.component';
import { NavigatorComponent } from '../navigator/navigator.component';
import { SessionService } from '../../shared/session.service';
import { AboutDialogComponent } from '../../shared/about-dialog/about-dialog.component';
import { SearchTriggerService } from '../global-search/search-trigger.service';
import { CommonRoutes } from "../../../lib/entities/shared.const";
import { ConfigScannerService, SCANNERS_DOC } from "../../config/scanner/config-scanner.service";
import { THEME_ARRAY, ThemeInterface } from "../../services/theme";
import { clone } from "../../../lib/utils/utils";
import { ThemeService } from "../../services/theme.service";

const HAS_SHOWED_SCANNER_INFO: string = 'hasShowScannerInfo';
const YES: string = 'yes';
const HAS_STYLE_MODE: string = 'styleModeLocal';

@Component({
    selector: 'harbor-shell',
    templateUrl: 'harbor-shell.component.html',
    styleUrls: ["harbor-shell.component.scss"]
})

export class HarborShellComponent implements OnInit, OnDestroy {

    @ViewChild(AccountSettingsModalComponent, { static: false })
    accountSettingsModal: AccountSettingsModalComponent;

    @ViewChild(PasswordSettingComponent, { static: false })
    pwdSetting: PasswordSettingComponent;

    @ViewChild(NavigatorComponent, { static: false })
    navigator: NavigatorComponent;

    @ViewChild(AboutDialogComponent, { static: false })
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
    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private session: SessionService,
        private searchTrigger: SearchTriggerService,
        private appConfigService: AppConfigService,
        private scannerService: ConfigScannerService,
        public theme: ThemeService,
    ) { }

    ngOnInit() {
        if (this.appConfigService.isLdapMode()) {
            this.isLdapMode = true;
        } else if (this.appConfigService.isHttpAuthMode()) {
            this.isHttpAuthMode = true;
        } else if (this.appConfigService.isOidcMode()) {
            this.isOidcMode = true;
        }
        this.searchSub = this.searchTrigger.searchTriggerChan$.subscribe(searchEvt => {
            if (searchEvt && searchEvt.trim() !== "") {
                this.isSearchResultsOpened = true;
            }
        });

        this.searchCloseSub = this.searchTrigger.searchCloseChan$.subscribe(close => {
            this.isSearchResultsOpened = false;
        });
        if (!(localStorage && localStorage.getItem(HAS_SHOWED_SCANNER_INFO) === YES)) {
            this.getDefaultScanner();
        }
        // set local in app
        if (localStorage) {
            this.styleMode = localStorage.getItem(HAS_STYLE_MODE);
        }
    }
    closeInfo() {
        if (localStorage) {
            localStorage.setItem(HAS_SHOWED_SCANNER_INFO, YES);
        }
        this.showScannerInfo = false;
    }

    getDefaultScanner() {
        this.scannerService.getScanners()
            .subscribe(scanners => {
                if (scanners && scanners.length) {
                    this.showScannerInfo = scanners.some(scanner => scanner.is_default);
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
    }

    public get shouldOverrideContent(): boolean {
        return this.router.routerState.snapshot.url.toString().startsWith(CommonRoutes.EMBEDDED_SIGN_IN);
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
        return this.session.getCurrentUser() &&
            this.session.getCurrentUser().has_admin_role;
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
