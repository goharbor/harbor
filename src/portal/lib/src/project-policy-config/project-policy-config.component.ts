import {Component, ElementRef, Input, OnInit, ViewChild} from '@angular/core';

import {compareValue, clone} from '../utils';
import {ProjectService} from '../service/project.service';
import {ErrorHandler} from '../error-handler/error-handler';
import {State, SystemCVEWhitelist} from '../service/interface';

import {ConfirmationState, ConfirmationTargets} from '../shared/shared.const';
import {ConfirmationMessage} from '../confirmation-dialog/confirmation-message';
import {ConfirmationDialogComponent} from '../confirmation-dialog/confirmation-dialog.component';
import {ConfirmationAcknowledgement} from '../confirmation-dialog/confirmation-state-message';
import {TranslateService} from '@ngx-translate/core';

import {Project} from './project';
import {SystemInfo, SystemInfoService} from '../service/index';
import {UserPermissionService} from '../service/permission.service';
import {USERSTATICPERMISSION} from '../service/permission-static';


const ONE_THOUSAND: number = 1000;
const LOW: string = 'low';
const CVE_DETAIL_PRE_URL = `https://nvd.nist.gov/vuln/detail/`;
const TARGET_BLANK = "_blank";

export class ProjectPolicy {
    Public: boolean;
    ContentTrust: boolean;
    PreventVulImg: boolean;
    PreventVulImgSeverity: string;
    ScanImgOnPush: boolean;

    constructor() {
        this.Public = false;
        this.ContentTrust = false;
        this.PreventVulImg = false;
        this.PreventVulImgSeverity = LOW;
        this.ScanImgOnPush = false;
    }

    initByProject(pro: Project) {
        this.Public = pro.metadata.public === 'true' ? true : false;
        this.ContentTrust = pro.metadata.enable_content_trust === 'true' ? true : false;
        this.PreventVulImg = pro.metadata.prevent_vul === 'true' ? true : false;
        if (pro.metadata.severity) {
            this.PreventVulImgSeverity = pro.metadata.severity;
        }
        this.ScanImgOnPush = pro.metadata.auto_scan === 'true' ? true : false;
    }
}

@Component({
    selector: 'hbr-project-policy-config',
    templateUrl: './project-policy-config.component.html',
    styleUrls: ['./project-policy-config.component.scss']
})
export class ProjectPolicyConfigComponent implements OnInit {
    onGoing = false;
    @Input() projectId: number;
    @Input() projectName = 'unknown';

    @Input() hasSignedIn: boolean;
    @Input() hasProjectAdminRole: boolean;

    @ViewChild('cfgConfirmationDialog') confirmationDlg: ConfirmationDialogComponent;
    @ViewChild('dateInput') dateInput: ElementRef;
    @ViewChild('dateSystemInput') dateSystemInput: ElementRef;

    systemInfo: SystemInfo;
    orgProjectPolicy = new ProjectPolicy();
    projectPolicy = new ProjectPolicy();
    hasChangeConfigRole: boolean;
    severityOptions = [
        {severity: 'high', severityLevel: 'VULNERABILITY.SEVERITY.HIGH'},
        {severity: 'medium', severityLevel: 'VULNERABILITY.SEVERITY.MEDIUM'},
        {severity: 'low', severityLevel: 'VULNERABILITY.SEVERITY.LOW'},
        {severity: 'negligible', severityLevel: 'VULNERABILITY.SEVERITY.NEGLIGIBLE'},
    ];
    userSystemWhitelist: boolean = true;
    showAddModal: boolean = false;
    systemWhitelist: SystemCVEWhitelist;
    cveIds: string;
    systemExpiresDate: Date;
    systemExpiresDateString: string;
    userProjectWhitelist = false;
    systemWhitelistOrProjectWhitelist: string;
    systemWhitelistOrProjectWhitelistOrigin: string;
    projectWhitelist;
    projectWhitelistOrigin;

    constructor(
        private errorHandler: ErrorHandler,
        private translate: TranslateService,
        private projectService: ProjectService,
        private systemInfoService: SystemInfoService,
        private userPermission: UserPermissionService,
    ) {
    }

    ngOnInit(): void {
        // assert if project id exist
        if (!this.projectId) {
            this.errorHandler.error('Project ID cannot be unset.');
            return;
        }
        // get system info
        this.systemInfoService.getSystemInfo()
            .subscribe(systemInfo => {
                this.systemInfo = systemInfo;
                if (this.withClair) {
                    setTimeout(() => {
                        this.dateSystemInput.nativeElement.parentNode.setAttribute("hidden", "hidden");
                    }, 100);
                }
            } , error => this.errorHandler.error(error));
        // retrive project level policy data
        this.retrieve();
        this.getPermission();
        this.getSystemWhitelist();
    }

    getSystemWhitelist() {
        this.systemInfoService.getSystemWhitelist()
            .subscribe((systemWhitelist) => {
                    if (systemWhitelist) {
                        this.systemWhitelist = systemWhitelist;
                        if (this.systemWhitelist.expires_at) {
                            this.systemExpiresDate = new Date(this.systemWhitelist.expires_at * ONE_THOUSAND);
                            setTimeout( () => {
                                this.systemExpiresDateString = this.dateSystemInput.nativeElement.value;
                            }, 100);
                        }
                    }
                }, error => {
                    this.errorHandler.error(error);
                }
            );
    }

    private getPermission(): void {
        this.userPermission.getPermission(this.projectId,
            USERSTATICPERMISSION.CONFIGURATION.KEY, USERSTATICPERMISSION.CONFIGURATION.VALUE.UPDATE).subscribe(permissins => {
            this.hasChangeConfigRole = permissins as boolean;
        });
    }

    public get withNotary(): boolean {
        return this.systemInfo ? this.systemInfo.with_notary : false;
    }

    public get withClair(): boolean {
        return this.systemInfo ? this.systemInfo.with_clair : false;
    }

    retrieve(state?: State): any {
        this.projectService.getProject(this.projectId)
            .subscribe(
                response => {
                    this.orgProjectPolicy.initByProject(response);
                    this.projectPolicy.initByProject(response);
                    // get projectWhitelist
                    if (!response.cve_whitelist) {
                       response.cve_whitelist = {
                            items: [],
                            expires_at: null
                        };
                    }
                    if (!response.cve_whitelist['items']) {
                        response.cve_whitelist['items'] = [];
                    }
                    if (!response.cve_whitelist['expires_at']) {
                        response.cve_whitelist['expires_at'] = null;
                    }
                    if (!response.metadata.reuse_sys_cve_whitelist) {
                        response.metadata.reuse_sys_cve_whitelist = "true";
                    }
                    if (response && response.cve_whitelist) {
                        this.projectWhitelist = clone(response.cve_whitelist);
                        this.projectWhitelistOrigin = clone(response.cve_whitelist);
                        this.systemWhitelistOrProjectWhitelist = response.metadata.reuse_sys_cve_whitelist;
                        this.systemWhitelistOrProjectWhitelistOrigin = response.metadata.reuse_sys_cve_whitelist;
                    }
                }, error => this.errorHandler.error(error));
    }

    refresh() {
        this.retrieve();
    }

    isValid() {
        let flag = false;
        if (!this.projectPolicy.PreventVulImg || this.severityOptions.some(x => x.severity === this.projectPolicy.PreventVulImgSeverity)) {
            flag = true;
        }
        return flag;
    }

    hasChanges() {
        return !compareValue(this.orgProjectPolicy, this.projectPolicy);
    }

    save() {
        if (!this.hasChanges() && !this.hasWhitelistChanged) {
            return;
        }
        this.onGoing = true;
        this.projectService.updateProjectPolicy(
            this.projectId,
            this.projectPolicy,
            this.systemWhitelistOrProjectWhitelist,
            this.projectWhitelist)
            .subscribe(() => {
                this.onGoing = false;
                this.translate.get('CONFIG.SAVE_SUCCESS').subscribe((res: string) => {
                    this.errorHandler.info(res);
                });
                this.refresh();
            }, error => {
                this.onGoing = false;
                this.errorHandler.error(error);
            });
    }

    cancel(): void {
        let msg = new ConfirmationMessage(
            'CONFIG.CONFIRM_TITLE',
            'CONFIG.CONFIRM_SUMMARY',
            '',
            {},
            ConfirmationTargets.CONFIG
        );
        this.confirmationDlg.open(msg);
    }

    reset(): void {
        this.projectPolicy = clone(this.orgProjectPolicy);
    }

    confirmCancel(ack: ConfirmationAcknowledgement): void {
        if (ack && ack.source === ConfirmationTargets.CONFIG &&
            ack.state === ConfirmationState.CONFIRMED) {
            this.reset();
            if (this.hasWhitelistChanged) {
                this.projectWhitelist = clone(this.projectWhitelistOrigin);
                this.systemWhitelistOrProjectWhitelist = this.systemWhitelistOrProjectWhitelistOrigin;
            }
        }
    }

    isUseSystemWhitelist(): boolean {
        return this.systemWhitelistOrProjectWhitelist === 'true';
    }

    deleteItem(index: number) {
        this.projectWhitelist.items.splice(index, 1);
    }

    addSystem() {
        this.showAddModal = false;
        if (!(this.systemWhitelist && this.systemWhitelist.items && this.systemWhitelist.items.length > 0)) {
            return;
        }
        if (this.projectWhitelist && !this.projectWhitelist.items) {
            this.projectWhitelist.items = [];
        }
        // remove duplication and add to projectWhitelist
        let map = {};
        this.projectWhitelist.items.forEach(item => {
            map[item.cve_id] = true;
        });
        this.systemWhitelist.items.forEach(item => {
            if (!map[item.cve_id]) {
                map[item.cve_id] = true;
                this.projectWhitelist.items.push(item);
            }
        });
    }

    addToProjectWhitelist() {
        if (this.projectWhitelist && !this.projectWhitelist.items) {
            this.projectWhitelist.items = [];
        }
        // remove duplication and add to projectWhitelist
        let map = {};
        this.projectWhitelist.items.forEach(item => {
            map[item.cve_id] = true;
        });
        this.cveIds.split(/[\n,]+/).forEach(id => {
            let cveObj: any = {};
            cveObj.cve_id = id.trim();
            if (!map[cveObj.cve_id]) {
                map[cveObj.cve_id] = true;
                this.projectWhitelist.items.push(cveObj);
            }
        });
        // clear modal and close modal
        this.cveIds = null;
        this.showAddModal = false;
    }

    get hasWhitelistChanged(): boolean {
        return !(compareValue(this.projectWhitelist, this.projectWhitelistOrigin)
            && this.systemWhitelistOrProjectWhitelistOrigin === this.systemWhitelistOrProjectWhitelist);
    }

    isDisabled(): boolean {
        let str = this.cveIds;
        return !(str && str.trim());
    }

    get expiresDate() {
        if (this.systemWhitelistOrProjectWhitelist === 'true') {
            if (this.systemWhitelist && this.systemWhitelist.expires_at) {
                return new Date(this.systemWhitelist.expires_at * ONE_THOUSAND);
            }
        } else {
            if (this.projectWhitelist && this.projectWhitelist.expires_at) {
                return new Date(this.projectWhitelist.expires_at * ONE_THOUSAND);
            }
        }
        return null;
    }

    set expiresDate(date) {
        if (this.systemWhitelistOrProjectWhitelist === 'false') {
            if (this.projectWhitelist && date) {
                this.projectWhitelist.expires_at = Math.floor(date.getTime() / ONE_THOUSAND);
            }
        }
    }

    get neverExpires(): boolean {
        if (this.systemWhitelistOrProjectWhitelist === 'true') {
            if (this.systemWhitelist && this.systemWhitelist.expires_at) {
                return !(this.systemWhitelist && this.systemWhitelist.expires_at);
            }
        } else {
            if (this.projectWhitelist && this.projectWhitelist.expires_at) {
                return !(this.projectWhitelist && this.projectWhitelist.expires_at);
            }
        }
        return true;
    }

    set neverExpires(flag) {
        if (flag) {
            this.projectWhitelist.expires_at = null;
            this.systemInfoService.resetDateInput(this.dateInput);
        } else {
            this.projectWhitelist.expires_at = Math.floor(new Date().getTime() / ONE_THOUSAND);
        }
    }

    get hasExpired(): boolean {
        if (this.systemWhitelistOrProjectWhitelist === 'true') {
            if (this.systemWhitelist && this.systemWhitelist.expires_at) {
                return new Date().getTime() > this.systemWhitelist.expires_at * ONE_THOUSAND;
            }
        } else {
            if (this.projectWhitelistOrigin && this.projectWhitelistOrigin.expires_at) {
                return new Date().getTime() > this.projectWhitelistOrigin.expires_at * ONE_THOUSAND;
            }
        }
        return false;
    }
    goToDetail(cveId) {
        window.open(CVE_DETAIL_PRE_URL + `${cveId}`, TARGET_BLANK);
    }
}
