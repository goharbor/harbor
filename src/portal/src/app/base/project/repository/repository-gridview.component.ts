import {
    ChangeDetectorRef,
    Component,
    EventEmitter,
    Input,
    OnChanges,
    OnDestroy,
    OnInit,
    Output,
    SimpleChanges,
    ViewChild,
} from '@angular/core';
import { forkJoin, Observable, of, Subscription } from 'rxjs';
import {
    catchError,
    debounceTime,
    distinctUntilChanged,
    finalize,
    map,
    switchMap,
} from 'rxjs/operators';
import { TranslateService } from '@ngx-translate/core';
import { ClrDatagridStateInterface } from '@clr/angular';
import { RepositoryService as NewRepositoryService } from '../../../../../ng-swagger-gen/services/repository.service';
import {
    SystemInfo,
    SystemInfoService,
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../shared/services';
import { FilterComponent } from '../../../shared/components/filter/filter.component';
import {
    calculatePage,
    clone,
    CURRENT_BASE_HREF,
    dbEncodeURIComponent,
    doFiltering,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../shared/units/utils';
import { ErrorHandler } from '../../../shared/units/error-handler';
import {
    CARD_VIEW_LOCALSTORAGE_KEY,
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
    FALSE_STR,
    TRUE_STR,
} from '../../../shared/entities/shared.const';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../shared/components/operation/operate';
import { ConfirmationDialogComponent } from '../../../shared/components/confirmation-dialog';
import { OperationService } from '../../../shared/components/operation/operation.service';
import { Project } from '../project';
import { ActivatedRoute, Router } from '@angular/router';
import { SessionService } from '../../../shared/services/session.service';
import { GridViewComponent } from './gridview/grid-view.component';
import { Repository as NewRepository } from '../../../../../ng-swagger-gen/models/repository';
import { StrictHttpResponse as __StrictHttpResponse } from '../../../../../ng-swagger-gen/strict-http-response';
import { HttpErrorResponse } from '@angular/common/http';
import { errorHandler } from '../../../shared/units/shared.utils';
import { ConfirmationAcknowledgement } from '../../global-confirmation-dialog/confirmation-state-message';
import { ConfirmationMessage } from '../../global-confirmation-dialog/confirmation-message';
import {
    EventService,
    HarborEvent,
} from '../../../services/event-service/event.service';

@Component({
    selector: 'hbr-repository-gridview',
    templateUrl: './repository-gridview.component.html',
    styleUrls: ['./repository-gridview.component.scss'],
})
export class RepositoryGridviewComponent
    implements OnChanges, OnInit, OnDestroy
{
    isFirstLoadingGridView: boolean = false;
    signedCon: { [key: string]: any | string[] } = {};
    downloadLink: string;
    @Input() urlPrefix: string;
    projectId: number;
    hasProjectAdminRole: boolean;
    hasSignedIn: boolean;
    projectName: string;
    mode = 'standalone';
    @Output() repoProvisionEvent = new EventEmitter<NewRepository>();
    @Output() addInfoEvent = new EventEmitter<NewRepository>();

    lastFilteredRepoName: string;
    repositories: NewRepository[] = [];
    repositoriesCopy: NewRepository[] = [];
    systemInfo: SystemInfo;
    selectedRow: NewRepository[] = [];
    loading = true;

    isCardView: boolean;
    cardHover = false;
    listHover = false;

    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.REPOSITORY_GRIDVIEW_COMPONENT
    );
    currentPage = 1;
    totalCount = 0;
    currentState: ClrDatagridStateInterface;

    @ViewChild('confirmationDialog')
    confirmationDialog: ConfirmationDialogComponent;

    @ViewChild('gridView') gridView: GridViewComponent;
    hasCreateRepositoryPermission: boolean;
    hasDeleteRepositoryPermission: boolean;
    @ViewChild(FilterComponent, { static: true })
    filterComponent: FilterComponent;
    searchSub: Subscription;
    isProxyCacheProject: boolean = false;

    constructor(
        private errorHandlerService: ErrorHandler,
        private translateService: TranslateService,
        private newRepoService: NewRepositoryService,
        private systemInfoService: SystemInfoService,
        private operationService: OperationService,
        private userPermissionService: UserPermissionService,
        private route: ActivatedRoute,
        private session: SessionService,
        private router: Router,
        private event: EventService,
        private cd: ChangeDetectorRef
    ) {
        if (localStorage) {
            this.isCardView =
                localStorage.getItem(CARD_VIEW_LOCALSTORAGE_KEY) === TRUE_STR;
        }
        this.downloadLink = CURRENT_BASE_HREF + '/systeminfo/getcert';
    }

    public get registryUrl(): string {
        return this.systemInfo ? this.systemInfo.registry_url : '';
    }
    public get withAdmiral(): boolean {
        return this.mode === 'admiral';
    }

    get canDownloadCert(): boolean {
        return this.systemInfo && this.systemInfo.has_ca_root;
    }

    getLink(repoEvt: NewRepository) {
        return [
            '/harbor/projects',
            repoEvt.project_id,
            'repositories',
            repoEvt.name.substr(this.projectName.length + 1),
        ];
    }

    ngOnChanges(changes: SimpleChanges): void {
        if (changes['projectId'] && changes['projectId'].currentValue) {
            this.refresh();
        }
    }

    ngOnInit(): void {
        this.projectId = this.route.snapshot.parent.parent.params['id'];
        let resolverData = this.route.snapshot.parent.parent.data;
        if (resolverData) {
            let pro: Project = <Project>resolverData['projectResolver'];
            this.hasProjectAdminRole = pro.has_project_admin_role;
            this.projectName = pro.name;
            if (pro.registry_id) {
                this.isProxyCacheProject = true;
            }
        }
        this.hasSignedIn = this.session.getCurrentUser() !== null;
        // Get system info for tag views
        this.getSystemInfo();
        if (this.isCardView) {
            this.doSearchRepoNames('', true);
        }
        this.lastFilteredRepoName = '';
        this.getHelmChartVersionPermission(this.projectId);
        if (!this.searchSub) {
            this.searchSub = this.filterComponent.filterTerms
                .pipe(
                    debounceTime(500),
                    distinctUntilChanged(),
                    switchMap(repoName => {
                        this.lastFilteredRepoName = repoName as string;
                        this.currentPage = 1;
                        // Pagination
                        let params: NewRepositoryService.ListRepositoriesParams =
                            {
                                projectName: this.projectName,
                                page: this.currentPage,
                                pageSize: this.pageSize,
                            };
                        if (this.lastFilteredRepoName) {
                            params.q = encodeURIComponent(
                                `name=~${this.lastFilteredRepoName}`
                            );
                        }
                        this.loading = true;
                        return this.newRepoService
                            .listRepositoriesResponse(params)
                            .pipe(finalize(() => (this.loading = false)));
                    })
                )
                .subscribe(
                    (repo: __StrictHttpResponse<Array<NewRepository>>) => {
                        this.totalCount = +repo.headers.get('x-total-count');
                        this.repositories = repo.body;
                    },
                    error => {
                        this.errorHandlerService.error(error);
                    }
                );
        }
    }
    getSystemInfo() {
        this.systemInfoService.getSystemInfo().subscribe(
            systemInfo => (this.systemInfo = systemInfo),
            error => this.errorHandlerService.error(error)
        );
    }
    ngOnDestroy() {
        if (this.searchSub) {
            this.searchSub.unsubscribe();
            this.searchSub = null;
        }
    }

    confirmDeletion(message: ConfirmationAcknowledgement) {
        if (
            message &&
            message.source === ConfirmationTargets.REPOSITORY &&
            message.state === ConfirmationState.CONFIRMED
        ) {
            let repoLists = message.data;
            if (repoLists && repoLists.length) {
                let observableLists: any[] = [];
                repoLists.forEach(repo => {
                    observableLists.push(this.delOperate(repo));
                });
                forkJoin(observableLists).subscribe(resArr => {
                    let error;
                    if (resArr && resArr.length) {
                        resArr.forEach(item => {
                            if (item instanceof HttpErrorResponse) {
                                error = errorHandler(item);
                            }
                        });
                    }
                    if (error) {
                        this.errorHandlerService.error(error);
                    } else {
                        this.translateService
                            .get('BATCH.DELETED_SUCCESS')
                            .subscribe(res => {
                                this.errorHandlerService.info(res);
                            });
                    }
                    this.selectedRow = [];
                    let st: ClrDatagridStateInterface =
                        this.getStateAfterDeletion();
                    if (!st) {
                        this.refresh();
                    } else {
                        this.clrLoad(st);
                    }
                });
            }
        }
    }

    delOperate(repo: NewRepository): Observable<any> {
        // init operation info
        let operMessage = new OperateInfo();
        operMessage.name = 'OPERATION.DELETE_REPO';
        operMessage.data.id = repo.id;
        operMessage.state = OperationState.progressing;
        repo.name = repo.name.substr(this.projectName.length + 1);
        operMessage.data.name = repo.name;

        this.operationService.publishInfo(operMessage);
        return this.newRepoService
            .deleteRepository({
                repositoryName: dbEncodeURIComponent(repo.name),
                projectName: this.projectName,
            })
            .pipe(
                map(response => {
                    this.translateService
                        .get('BATCH.DELETED_SUCCESS')
                        .subscribe(res => {
                            operateChanges(operMessage, OperationState.success);
                        });
                }),
                catchError(error => {
                    const message = errorHandler(error);
                    this.translateService.get(message).subscribe(res => {
                        operateChanges(
                            operMessage,
                            OperationState.failure,
                            res
                        );
                    });
                    return of(error);
                })
            );
    }

    doSearchRepoNames(repoName: string, isFirstLoadingGridView?: boolean) {
        this.lastFilteredRepoName = repoName;
        this.currentPage = 1;
        let st: ClrDatagridStateInterface = this.currentState;
        if (!st || !st.page) {
            st = { page: {} };
        }
        st.page.size = this.pageSize;
        st.page.from = 0;
        st.page.to = this.pageSize - 1;
        this.clrLoad(st, isFirstLoadingGridView);
    }

    deleteRepos(repoLists: NewRepository[]) {
        if (repoLists && repoLists.length) {
            let repoNames: string[] = [];
            repoLists.forEach(repo => {
                repoNames.push(repo.name);
            });
            this.confirmationDialogSet(
                'REPOSITORY.DELETION_TITLE_REPO',
                '',
                repoNames.join(','),
                repoLists,
                'REPOSITORY.DELETION_SUMMARY_REPO',
                ConfirmationButtons.DELETE_CANCEL
            );
        }
    }

    confirmationDialogSet(
        summaryTitle: string,
        signature: string,
        repoName: string,
        repoLists: NewRepository[],
        summaryKey: string,
        button: ConfirmationButtons
    ): void {
        this.translateService
            .get(summaryKey, {
                repoName: repoName,
                signedImages: signature,
            })
            .subscribe((res: string) => {
                summaryKey = res;
                let message = new ConfirmationMessage(
                    summaryTitle,
                    summaryKey,
                    repoName,
                    repoLists,
                    ConfirmationTargets.REPOSITORY,
                    button
                );
                this.confirmationDialog.open(message);
            });
    }

    itemAddInfoEvent(evt: any, repo: NewRepository): void {
        evt.stopPropagation();
        let repoCopy = clone(repo);
        repoCopy.name = this.registryUrl + ':443/' + repoCopy.name;
        this.addInfoEvent.emit(repoCopy);
    }

    deleteItemEvent(evt: any, item: NewRepository): void {
        evt.stopPropagation();
        this.deleteRepos([item]);
    }

    refresh() {
        this.doSearchRepoNames('');
        // notify project detail component to refresh project info
        this.event.publish(HarborEvent.REFRESH_PROJECT_INFO);
    }

    loadNextPage() {
        this.currentPage = this.currentPage + 1;
        // Pagination
        let params: NewRepositoryService.ListRepositoriesParams = {
            projectName: this.projectName,
            page: this.currentPage,
            pageSize: this.pageSize,
        };
        if (this.lastFilteredRepoName) {
            params.q = encodeURIComponent(`name=~${this.lastFilteredRepoName}`);
        }
        this.loading = true;
        this.newRepoService
            .listRepositoriesResponse(params)
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                (repo: __StrictHttpResponse<Array<NewRepository>>) => {
                    this.totalCount = +repo.headers.get('x-total-count');
                    this.repositoriesCopy = repo.body;
                    this.repositories = this.repositories.concat(
                        this.repositoriesCopy
                    );
                },
                error => {
                    this.errorHandlerService.error(error);
                }
            );
    }

    clrLoad(
        state: ClrDatagridStateInterface,
        isFirstLoadingGridView?: boolean
    ): void {
        if (!state || !state.page) {
            return;
        }
        this.pageSize = state.page.size;
        setPageSizeToLocalStorage(
            PageSizeMapKeys.REPOSITORY_GRIDVIEW_COMPONENT,
            this.pageSize
        );
        this.selectedRow = [];
        // Keep it for future filtering and sorting
        this.currentState = state;

        let pageNumber: number = calculatePage(state);
        if (pageNumber <= 0) {
            pageNumber = 1;
        }

        // Pagination
        let params: NewRepositoryService.ListRepositoriesParams = {
            projectName: this.projectName,
            page: pageNumber,
            pageSize: this.pageSize,
        };
        if (this.lastFilteredRepoName) {
            params.q = encodeURIComponent(`name=~${this.lastFilteredRepoName}`);
        }
        if (state.filters && state.filters.length) {
            state.filters.forEach(item => {
                params[item.property] = item.value;
            });
        }
        if (state.sort && state.sort.by) {
            params.sort = getSortingString(state);
        }
        this.loading = true;
        if (isFirstLoadingGridView) {
            this.isFirstLoadingGridView = true;
        }
        this.newRepoService
            .listRepositoriesResponse(params)
            .pipe(
                finalize(() => {
                    this.loading = false;
                    this.isFirstLoadingGridView = false;
                })
            )
            .subscribe(
                (repo: __StrictHttpResponse<Array<NewRepository>>) => {
                    this.totalCount = +repo.headers.get('x-total-count');
                    this.repositories = repo.body;
                    // Do customising filtering and sorting
                    this.repositories = doFiltering<NewRepository>(
                        this.repositories,
                        state
                    );
                    this.signedCon = {};
                },
                error => {
                    this.errorHandlerService.error(error);
                }
            );
    }

    getStateAfterDeletion(): ClrDatagridStateInterface {
        let total: number = this.totalCount - 1;
        if (total <= 0) {
            return null;
        }

        let totalPages: number = Math.ceil(total / this.pageSize);
        let targetPageNumber: number = this.currentPage;

        if (this.currentPage > totalPages) {
            targetPageNumber = totalPages; // Should == currentPage -1
        }

        let st: ClrDatagridStateInterface = this.currentState;
        if (!st) {
            st = { page: {} };
        }
        st.page.size = this.pageSize;
        st.page.from = (targetPageNumber - 1) * this.pageSize;
        st.page.to = targetPageNumber * this.pageSize - 1;

        return st;
    }
    getImgLink(repo: NewRepository): string {
        return '/container-image-icons?container-image=' + repo.name;
    }

    showCard(cardView: boolean) {
        if (this.isCardView === cardView) {
            return;
        }
        this.isCardView = cardView;
        // manually run change detecting to avoid ng-change-checking error
        this.cd.detectChanges();
        if (localStorage) {
            if (this.isCardView) {
                localStorage.setItem(CARD_VIEW_LOCALSTORAGE_KEY, TRUE_STR);
            } else {
                localStorage.setItem(CARD_VIEW_LOCALSTORAGE_KEY, FALSE_STR);
            }
        }
        if (this.isCardView) {
            this.refresh();
        }
    }

    mouseEnter(itemName: string) {
        if (itemName === 'card') {
            this.cardHover = true;
        } else {
            this.listHover = true;
        }
    }

    mouseLeave(itemName: string) {
        if (itemName === 'card') {
            this.cardHover = false;
        } else {
            this.listHover = false;
        }
    }

    isHovering(itemName: string) {
        if (itemName === 'card') {
            return this.cardHover;
        } else {
            return this.listHover;
        }
    }

    getHelmChartVersionPermission(projectId: number): void {
        let hasCreateRepositoryPermission =
            this.userPermissionService.getPermission(
                this.projectId,
                USERSTATICPERMISSION.REPOSITORY.KEY,
                USERSTATICPERMISSION.REPOSITORY.VALUE.CREATE
            );
        let hasDeleteRepositoryPermission =
            this.userPermissionService.getPermission(
                this.projectId,
                USERSTATICPERMISSION.REPOSITORY.KEY,
                USERSTATICPERMISSION.REPOSITORY.VALUE.DELETE
            );
        forkJoin(
            hasCreateRepositoryPermission,
            hasDeleteRepositoryPermission
        ).subscribe(
            permissions => {
                this.hasCreateRepositoryPermission = permissions[0] as boolean;
                this.hasDeleteRepositoryPermission = permissions[1] as boolean;
            },
            error => this.errorHandlerService.error(error)
        );
    }
}
