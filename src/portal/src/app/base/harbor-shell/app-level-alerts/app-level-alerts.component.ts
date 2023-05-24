import { Component, OnDestroy, OnInit } from '@angular/core';
import { SCANNERS_DOC } from '../../left-side-nav/interrogation-services/scanner/scanner';
import { SessionService } from '../../../shared/services/session.service';
import { DEFAULT_PAGE_SIZE, delUrlParam } from '../../../shared/units/utils';
import { forkJoin, Observable, Subscription } from 'rxjs';
import { Project } from '../../../../../ng-swagger-gen/models/project';
import { ScannerService } from '../../../../../ng-swagger-gen/services/scanner.service';
import { UN_LOGGED_PARAM } from '../../../account/sign-in/sign-in.service';
import {
    CommonRoutes,
    httpStatusCode,
} from '../../../shared/entities/shared.const';
import { ActivatedRoute, Router } from '@angular/router';
import { MessageService } from '../../../shared/components/global-message/message.service';
import { Message } from '../../../shared/components/global-message/message';
import { JobServiceDashboardHealthCheckService } from '../../left-side-nav/job-service-dashboard/job-service-dashboard-health-check.service';
import { AppConfigService } from '../../../services/app-config.service';
const HAS_SHOWED_SCANNER_INFO: string = 'hasShowScannerInfo';
const YES: string = 'yes';
@Component({
    selector: 'app-app-level-alerts',
    templateUrl: './app-level-alerts.component.html',
    styleUrls: ['./app-level-alerts.component.scss'],
})
export class AppLevelAlertsComponent implements OnInit, OnDestroy {
    scannerDocUrl: string = SCANNERS_DOC;
    showScannerInfo: boolean = false;
    message: Message;
    appLevelMsgSub: Subscription;
    clearSub: Subscription;
    showLogin: boolean = false;
    constructor(
        private session: SessionService,
        private scannerService: ScannerService,
        private router: Router,
        private messageService: MessageService,
        private route: ActivatedRoute,
        private jobServiceDashboardHealthCheckService: JobServiceDashboardHealthCheckService,
        private appConfigService: AppConfigService
    ) {}
    ngOnInit() {
        if (
            !(
                localStorage &&
                localStorage.getItem(HAS_SHOWED_SCANNER_INFO) === YES
            )
        ) {
            if (this.session.getCurrentUser()?.has_admin_role) {
                this.getDefaultScanner();
            }
        }
        if (!this.appLevelMsgSub) {
            this.appLevelMsgSub =
                this.messageService.appLevelAnnounced$.subscribe(message => {
                    this.message = message;
                    if (message.statusCode === httpStatusCode.Unauthorized) {
                        this.showLogin = true;
                        // User session timed out, then redirect to sign-in page
                        if (
                            this.session.getCurrentUser() &&
                            !this.isSignInUrl() &&
                            this.route.snapshot.queryParams[UN_LOGGED_PARAM] !==
                                YES
                        ) {
                            const url = delUrlParam(
                                this.router.url,
                                UN_LOGGED_PARAM
                            );
                            this.session.clear(); // because of SignInGuard, must clear user session before navigating to sign-in page
                            this.router.navigate(
                                [CommonRoutes.EMBEDDED_SIGN_IN],
                                {
                                    queryParams: { redirect_url: url },
                                }
                            );
                        }
                    } else {
                        this.showLogin = false;
                    }
                });
        }
        if (!this.clearSub) {
            this.clearSub = this.messageService.clearChan$.subscribe(clear => {
                this.showLogin = false;
            });
        }
    }
    ngOnDestroy() {
        if (this.appLevelMsgSub) {
            this.appLevelMsgSub.unsubscribe();
            this.appLevelMsgSub = null;
        }
    }
    get showReadOnly(): boolean {
        return this.appConfigService.getConfig()?.read_only;
    }
    shouldShowScannerInfo(): boolean {
        return (
            this.session.getCurrentUser()?.has_admin_role &&
            this.showScannerInfo
        );
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

    closeInfo() {
        if (localStorage) {
            localStorage.setItem(HAS_SHOWED_SCANNER_INFO, YES);
        }
        this.showScannerInfo = false;
    }

    signIn(): void {
        // remove queryParam UN_LOGGED_PARAM of redirect url
        const url = delUrlParam(this.router.url, UN_LOGGED_PARAM);
        this.router.navigate([CommonRoutes.EMBEDDED_SIGN_IN], {
            queryParams: { redirect_url: url },
        });
    }

    isSignInUrl(): boolean {
        const url: string =
            this.router.url?.indexOf('?') === -1
                ? this.router.url
                : this.router.url?.split('?')[0];
        return url === CommonRoutes.EMBEDDED_SIGN_IN;
    }

    showJobServiceDashboardHealthCheck(): boolean {
        return (
            this.jobServiceDashboardHealthCheckService.hasUnhealthyQueue() &&
            !this.jobServiceDashboardHealthCheckService.hasManuallyClosed()
        );
    }

    closeHealthWarning() {
        this.jobServiceDashboardHealthCheckService.setManuallyClosed(true);
    }

    isLogin(): boolean {
        return !!this.session.getCurrentUser();
    }
}
