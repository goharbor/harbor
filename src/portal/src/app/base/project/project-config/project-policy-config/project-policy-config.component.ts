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
import { Component, ElementRef, Input, OnInit, ViewChild } from '@angular/core';
import { compareValue, clone } from '../../../../shared/units/utils';
import {
    ProjectCVEAllowlist,
    ProjectService,
} from '../../../../shared/services';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { State, SystemCVEAllowlist } from '../../../../shared/services';
import {
    BandwidthUnit,
    ConfirmationState,
    ConfirmationTargets,
} from '../../../../shared/entities/shared.const';
import { ConfirmationMessage } from '../../../global-confirmation-dialog/confirmation-message';
import { ConfirmationDialogComponent } from '../../../../shared/components/confirmation-dialog';
import { ConfirmationAcknowledgement } from '../../../global-confirmation-dialog/confirmation-state-message';
import { TranslateService } from '@ngx-translate/core';
import { Project } from './project';
import { SystemInfo, SystemInfoService } from '../../../../shared/services';
import { UserPermissionService } from '../../../../shared/services';
import { USERSTATICPERMISSION } from '../../../../shared/services';
import { SessionService } from '../../../../shared/services/session.service';
import { Registry } from '../../../../../../ng-swagger-gen/models/registry';
import {
    EventService,
    HarborEvent,
} from '../../../../services/event-service/event.service';
import { forkJoin, Observable } from 'rxjs';
import { MessageHandlerService } from 'src/app/shared/services/message-handler.service';
import { RegistryService } from 'ng-swagger-gen/services';

const ONE_THOUSAND: number = 1000;
const LOW: string = 'low';
const CVE_DETAIL_PRE_URL = `https://nvd.nist.gov/vuln/detail/`;
const TARGET_BLANK = '_blank';

export class ProjectPolicy {
    Public: boolean;
    ContentTrust: boolean;
    ContentTrustCosign: boolean;
    PreventVulImg: boolean;
    PreventVulImgSeverity: string;
    ScanImgOnPush: boolean;
    GenerateSbomOnPush: boolean;
    ProxyCacheEnabled: boolean;
    RegistryId?: number | null;
    ProxySpeedKb?: number | null;
    MaxUpstreamConn?: number | null;

    constructor() {
        this.Public = false;
        this.ContentTrust = false;
        this.ContentTrustCosign = false;
        this.PreventVulImg = false;
        this.PreventVulImgSeverity = LOW;
        this.ScanImgOnPush = false;
        this.GenerateSbomOnPush = false;
        this.ProxyCacheEnabled = false;
        this.RegistryId = null;
        this.ProxySpeedKb = -1;
        this.MaxUpstreamConn = -1;
    }

    initByProject(pro: Project) {
        this.Public = pro.metadata.public === 'true';
        this.ContentTrust = pro.metadata.enable_content_trust === 'true';
        this.ContentTrustCosign =
            pro.metadata.enable_content_trust_cosign === 'true';
        this.PreventVulImg = pro.metadata.prevent_vul === 'true';
        if (pro.metadata.severity) {
            this.PreventVulImgSeverity = pro.metadata.severity;
        }
        this.ScanImgOnPush = pro.metadata.auto_scan === 'true';
        this.GenerateSbomOnPush = pro.metadata.auto_sbom_generation === 'true';
        this.ProxyCacheEnabled = pro.registry_id ? true : false;
        this.RegistryId = pro.registry_id;
        this.ProxySpeedKb = pro.metadata.proxy_speed_kb
            ? pro.metadata.proxy_speed_kb
            : -1;
        this.MaxUpstreamConn = pro.metadata.max_upstream_conn
            ? pro.metadata.max_upstream_conn
            : -1;
    }
}
const PAGE_SIZE: number = 100;

@Component({
    selector: 'hbr-project-policy-config',
    templateUrl: './project-policy-config.component.html',
    styleUrls: ['./project-policy-config.component.scss'],
})
export class ProjectPolicyConfigComponent implements OnInit {
    onGoing = false;
    allowUpdateProxyCacheConfiguration = false;
    @Input() projectId: number;
    @Input() projectName = 'unknown';
    @Input() isProxyCacheProject: boolean = false;

    @Input() hasSignedIn: boolean;
    @Input() hasProjectAdminRole: boolean;

    @ViewChild('cfgConfirmationDialog')
    confirmationDlg: ConfirmationDialogComponent;
    @ViewChild('dateInput') dateInput: ElementRef;
    @ViewChild('dateSystemInput') dateSystemInput: ElementRef;

    systemInfo: SystemInfo;
    orgProjectPolicy = new ProjectPolicy();
    projectPolicy = new ProjectPolicy();
    hasChangeConfigRole: boolean;

    severityOptions = [
        {
            severity: 'critical',
            severityLevel: 'VULNERABILITY.SEVERITY.CRITICAL',
        },
        { severity: 'high', severityLevel: 'VULNERABILITY.SEVERITY.HIGH' },
        { severity: 'medium', severityLevel: 'VULNERABILITY.SEVERITY.MEDIUM' },
        { severity: 'low', severityLevel: 'VULNERABILITY.SEVERITY.LOW' },
        { severity: 'none', severityLevel: 'VULNERABILITY.SEVERITY.NONE' },
    ];
    userSystemAllowlist: boolean = true;
    showAddModal: boolean = false;
    systemAllowlist: SystemCVEAllowlist;
    cveIds: string;
    systemExpiresDate: Date;
    systemExpiresDateString: string;
    userProjectAllowlist = false;
    systemAllowlistOrProjectAllowlist: string;
    systemAllowlistOrProjectAllowlistOrigin: string;
    projectAllowlist: ProjectCVEAllowlist;
    projectAllowlistOrigin: ProjectCVEAllowlist;
    speedUnit = BandwidthUnit.KB;
    speedUnits = [
        {
            UNIT: BandwidthUnit.KB,
        },
        {
            UNIT: BandwidthUnit.MB,
        },
    ];
    // **Added property for bandwidth error message**
    bandwidthError: string | null = null;
    maxUpstreamConnError: string | null = null;
    registries: Registry[] = [];
    supportedRegistryTypeQueryString: string =
        'type={docker-hub harbor azure-acr aws-ecr google-gcr quay docker-registry github-ghcr jfrog-artifactory}';

    constructor(
        private errorHandler: ErrorHandler,
        private translate: TranslateService,
        private projectService: ProjectService,
        private systemInfoService: SystemInfoService,
        private userPermission: UserPermissionService,
        private session: SessionService,
        private messageHandlerService: MessageHandlerService,
        private endpointService: RegistryService,
        private event: EventService
    ) {}

    ngOnInit(): void {
        // assert if project id exist
        if (!this.projectId) {
            this.errorHandler.error('Project ID cannot be unset.');
            return;
        }
        // get system info
        this.systemInfoService.getSystemInfo().subscribe(
            systemInfo => {
                this.systemInfo = systemInfo;
                setTimeout(() => {
                    this.dateSystemInput.nativeElement.parentNode.setAttribute(
                        'hidden',
                        'hidden'
                    );
                }, 100);
            },
            error => this.errorHandler.error(error)
        );
        // retrive project level policy data
        this.retrieve();
        this.getPermission();
        this.getSystemAllowlist();
        if (this.isSystemAdmin) {
            this.getRegistries();
        }
    }

    validateBandwidth(): void {
        const value = Number(this.projectPolicy.ProxySpeedKb);
        if (
            isNaN(value) ||
            (!Number.isInteger(value) && value !== -1) ||
            (value <= 0 && value !== -1)
        ) {
            this.translate
                .get('PROJECT.SPEED_LIMIT_TIP')
                .subscribe((res: string) => {
                    this.bandwidthError = res;
                });
        } else {
            this.bandwidthError = null;
        }
    }

    getRegistries() {
        this.endpointService
            .listRegistriesResponse({
                page: 1,
                pageSize: PAGE_SIZE,
                q: this.supportedRegistryTypeQueryString,
            })
            .subscribe(
                result => {
                    // Get total count
                    if (result.headers) {
                        const xHeader: string =
                            result.headers.get('X-Total-Count');
                        const totalCount = parseInt(xHeader, 0);
                        let arr = result.body || [];
                        if (totalCount <= PAGE_SIZE) {
                            // already gotten all Registries
                            this.registries = result.body || [];
                        } else {
                            // get all the registries in specified times
                            const times: number = Math.ceil(
                                totalCount / PAGE_SIZE
                            );
                            const observableList: Observable<Registry[]>[] = [];
                            for (let i = 2; i <= times; i++) {
                                observableList.push(
                                    this.endpointService.listRegistries({
                                        page: i,
                                        pageSize: PAGE_SIZE,
                                        q: this
                                            .supportedRegistryTypeQueryString,
                                    })
                                );
                            }
                            forkJoin(observableList).subscribe(res => {
                                if (res && res.length) {
                                    res.forEach(item => {
                                        arr = arr.concat(item);
                                    });
                                    this.registries = arr;
                                }
                            });
                        }
                    }
                },
                error => {
                    this.messageHandlerService.error(error);
                }
            );
    }

    getSystemAllowlist() {
        this.systemInfoService.getSystemAllowlist().subscribe(
            systemAllowlist => {
                if (systemAllowlist) {
                    this.systemAllowlist = systemAllowlist;
                    if (this.systemAllowlist.expires_at) {
                        this.systemExpiresDate = new Date(
                            this.systemAllowlist.expires_at * ONE_THOUSAND
                        );
                        setTimeout(() => {
                            this.systemExpiresDateString =
                                this.dateSystemInput.nativeElement.value;
                        }, 100);
                    }
                }
            },
            error => {
                this.errorHandler.error(error);
            }
        );
    }

    private getPermission(): void {
        this.userPermission
            .getPermission(
                this.projectId,
                USERSTATICPERMISSION.CONFIGURATION.KEY,
                USERSTATICPERMISSION.CONFIGURATION.VALUE.UPDATE
            )
            .subscribe(permissins => {
                this.hasChangeConfigRole = permissins as boolean;
            });
    }

    public get isSystemAdmin(): boolean {
        let account = this.session.getCurrentUser();
        return account != null && account.has_admin_role;
    }

    retrieve(state?: State): any {
        this.projectService.getProject(this.projectId).subscribe(
            response => {
                this.orgProjectPolicy.initByProject(response);
                this.projectPolicy.initByProject(response);
                // get projectAllowlist
                if (!response.cve_allowlist) {
                    response.cve_allowlist = {
                        items: [],
                        expires_at: null,
                    };
                }
                if (!response.cve_allowlist['items']) {
                    response.cve_allowlist['items'] = [];
                }
                if (!response.cve_allowlist['expires_at']) {
                    response.cve_allowlist['expires_at'] = null;
                }
                if (!response.metadata.reuse_sys_cve_allowlist) {
                    response.metadata.reuse_sys_cve_allowlist = 'true';
                }
                if (response && response.cve_allowlist) {
                    this.projectAllowlist = clone(response.cve_allowlist);
                    this.projectAllowlistOrigin = clone(response.cve_allowlist);
                    this.systemAllowlistOrProjectAllowlist =
                        response.metadata.reuse_sys_cve_allowlist;
                    this.systemAllowlistOrProjectAllowlistOrigin =
                        response.metadata.reuse_sys_cve_allowlist;
                }
            },
            error => this.errorHandler.error(error)
        );
    }

    refresh() {
        this.retrieve();
        // notify project detail component to refresh project info
        this.event.publish(HarborEvent.REFRESH_PROJECT_INFO);
    }

    isValid() {
        let flag = false;
        if (
            !this.projectPolicy.PreventVulImg ||
            this.severityOptions.some(
                x => x.severity === this.projectPolicy.PreventVulImgSeverity
            )
        ) {
            flag = true;
        }
        return flag;
    }

    hasChanges() {
        return !compareValue(this.orgProjectPolicy, this.projectPolicy);
    }

    save() {
        if (!this.hasChanges() && !this.hasAllowlistChanged) {
            return;
        }
        this.onGoing = true;
        this.projectService
            .updateProjectPolicy(
                this.projectId,
                this.projectPolicy,
                this.systemAllowlistOrProjectAllowlist,
                this.projectAllowlist
            )
            .subscribe(
                () => {
                    this.onGoing = false;
                    this.translate
                        .get('CONFIG.SAVE_SUCCESS')
                        .subscribe((res: string) => {
                            this.errorHandler.info(res);
                        });
                    this.refresh();
                },
                error => {
                    this.onGoing = false;
                    this.errorHandler.error(error);
                }
            );
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
        if (
            ack &&
            ack.source === ConfirmationTargets.CONFIG &&
            ack.state === ConfirmationState.CONFIRMED
        ) {
            this.reset();
            if (this.hasAllowlistChanged) {
                this.projectAllowlist = clone(this.projectAllowlistOrigin);
                this.systemAllowlistOrProjectAllowlist =
                    this.systemAllowlistOrProjectAllowlistOrigin;
            }
        }
    }

    isUseSystemAllowlist(): boolean {
        return this.systemAllowlistOrProjectAllowlist === 'true';
    }

    deleteItem(index: number) {
        this.projectAllowlist.items.splice(index, 1);
    }

    addSystem() {
        this.showAddModal = false;
        if (
            !(
                this.systemAllowlist &&
                this.systemAllowlist.items &&
                this.systemAllowlist.items.length > 0
            )
        ) {
            return;
        }
        if (this.projectAllowlist && !this.projectAllowlist.items) {
            this.projectAllowlist.items = [];
        }
        // remove duplication and add to projectAllowlist
        let map = {};
        this.projectAllowlist.items.forEach(item => {
            map[item.cve_id] = true;
        });
        this.systemAllowlist.items.forEach(item => {
            if (!map[item.cve_id]) {
                map[item.cve_id] = true;
                this.projectAllowlist.items.push(item);
            }
        });
    }

    addToProjectAllowlist() {
        if (this.projectAllowlist && !this.projectAllowlist.items) {
            this.projectAllowlist.items = [];
        }
        // remove duplication and add to projectAllowlist
        let map = {};
        this.projectAllowlist.items.forEach(item => {
            map[item.cve_id] = true;
        });
        const newCveIds = this.cveIds
            .split(/[\n,]+/)
            .map(id => id.trim()) // remove leading/trailing whitespace
            .filter(id => id.length > 0); // skip empty or whitespace-only strings

        newCveIds.forEach(id => {
            let cveObj: any = {};
            cveObj.cve_id = id.trim();
            if (!map[cveObj.cve_id]) {
                map[cveObj.cve_id] = true;
                this.projectAllowlist.items.push(cveObj);
            }
        });
        // clear modal and close modal
        this.cveIds = null;
        this.showAddModal = false;
    }

    get hasAllowlistChanged(): boolean {
        return !(
            compareValue(this.projectAllowlist, this.projectAllowlistOrigin) &&
            this.systemAllowlistOrProjectAllowlistOrigin ===
                this.systemAllowlistOrProjectAllowlist
        );
    }

    isDisabled(): boolean {
        let str = this.cveIds;
        return !(str && str.trim());
    }

    get expiresDate() {
        if (this.systemAllowlistOrProjectAllowlist === 'true') {
            if (this.systemAllowlist && this.systemAllowlist.expires_at) {
                return new Date(this.systemAllowlist.expires_at * ONE_THOUSAND);
            }
        } else {
            if (this.projectAllowlist && this.projectAllowlist.expires_at) {
                return new Date(
                    this.projectAllowlist.expires_at * ONE_THOUSAND
                );
            }
        }
        return null;
    }

    set expiresDate(date) {
        if (this.systemAllowlistOrProjectAllowlist === 'false') {
            if (this.projectAllowlist && date) {
                this.projectAllowlist.expires_at = Math.floor(
                    date.getTime() / ONE_THOUSAND
                );
            }
        }
    }

    get neverExpires(): boolean {
        if (this.systemAllowlistOrProjectAllowlist === 'true') {
            if (this.systemAllowlist && this.systemAllowlist.expires_at) {
                return !(
                    this.systemAllowlist && this.systemAllowlist.expires_at
                );
            }
        } else {
            if (this.projectAllowlist && this.projectAllowlist.expires_at) {
                return !(
                    this.projectAllowlist && this.projectAllowlist.expires_at
                );
            }
        }
        return true;
    }

    set neverExpires(flag) {
        if (flag) {
            this.projectAllowlist.expires_at = null;
            this.systemInfoService.resetDateInput(this.dateInput);
        } else {
            this.projectAllowlist.expires_at = Math.floor(
                new Date().getTime() / ONE_THOUSAND
            );
        }
    }

    get hasExpired(): boolean {
        if (this.systemAllowlistOrProjectAllowlist === 'true') {
            if (this.systemAllowlist && this.systemAllowlist.expires_at) {
                return (
                    new Date().getTime() >
                    this.systemAllowlist.expires_at * ONE_THOUSAND
                );
            }
        } else {
            if (
                this.projectAllowlistOrigin &&
                this.projectAllowlistOrigin.expires_at
            ) {
                return (
                    new Date().getTime() >
                    this.projectAllowlistOrigin.expires_at * ONE_THOUSAND
                );
            }
        }
        return false;
    }
    goToDetail(cveId) {
        window.open(CVE_DETAIL_PRE_URL + `${cveId}`, TARGET_BLANK);
    }
}
