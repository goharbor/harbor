import {
    Component,
    Input,
    Output,
    OnInit,
    ViewChild,
    ChangeDetectionStrategy,
    ChangeDetectorRef,
    EventEmitter,
    OnChanges,
    SimpleChanges
} from "@angular/core";
import { Router } from "@angular/router";
import { forkJoin } from "rxjs";
import { finalize } from "rxjs/operators";
import {TranslateService} from "@ngx-translate/core";
import {Comparator, State} from "@clr/angular";

import {
    Repository,
    SystemInfo,
    SystemInfoService,
    RepositoryService,
    RequestQueryParams,
    RepositoryItem,
    TagService
} from '../service/index';
import {ErrorHandler} from '../error-handler/error-handler';
import {toPromise, CustomComparator, DEFAULT_PAGE_SIZE, calculatePage, doFiltering, doSorting, clone} from '../utils';
import {ConfirmationState, ConfirmationTargets, ConfirmationButtons} from '../shared/shared.const';
import {ConfirmationDialogComponent} from '../confirmation-dialog/confirmation-dialog.component';
import {ConfirmationMessage} from '../confirmation-dialog/confirmation-message';
import {ConfirmationAcknowledgement} from '../confirmation-dialog/confirmation-state-message';
import {Tag} from '../service/interface';
import {GridViewComponent} from '../gridview/grid-view.component';
import {OperationService} from "../operation/operation.service";
import {OperateInfo, OperationState, operateChanges} from "../operation/operate";
import { downloadUrl } from '../service.config';
@Component({
    selector: "hbr-repository-gridview",
    templateUrl: "./repository-gridview.component.html",
    styleUrls: ["./repository-gridview.component.scss"],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class RepositoryGridviewComponent implements OnChanges, OnInit {
    signedCon: { [key: string]: any | string[] } = {};
    downloadLink: string = downloadUrl;
    @Input() projectId: number;
    @Input() projectName = "unknown";
    @Input() urlPrefix: string;
    @Input() hasSignedIn: boolean;
    @Input() hasProjectAdminRole: boolean;
    @Input() mode = "admiral";
    @Output() repoClickEvent = new EventEmitter<RepositoryItem>();
    @Output() repoProvisionEvent = new EventEmitter<RepositoryItem>();
    @Output() addInfoEvent = new EventEmitter<RepositoryItem>();

    lastFilteredRepoName: string;
    repositories: RepositoryItem[] = [];
    repositoriesCopy: RepositoryItem[] = [];
    systemInfo: SystemInfo;
    selectedRow: RepositoryItem[] = [];
    loading = true;

    isCardView: boolean;
    cardHover = false;
    listHover = false;

    pullCountComparator: Comparator<RepositoryItem> = new CustomComparator<RepositoryItem>('pull_count', 'number');
    tagsCountComparator: Comparator<RepositoryItem> = new CustomComparator<RepositoryItem>('tags_count', 'number');

    pageSize: number = DEFAULT_PAGE_SIZE;
    currentPage = 1;
    totalCount = 0;
    currentState: State;

    @ViewChild("confirmationDialog")
    confirmationDialog: ConfirmationDialogComponent;

    @ViewChild("gridView") gridView: GridViewComponent;

    constructor(private errorHandler: ErrorHandler,
                private translateService: TranslateService,
                private repositoryService: RepositoryService,
                private systemInfoService: SystemInfoService,
                private tagService: TagService,
                private operationService: OperationService,
                private ref: ChangeDetectorRef,
                private router: Router) {
    }

    public get registryUrl(): string {
        return this.systemInfo ? this.systemInfo.registry_url : "";
    }

    public get withClair(): boolean {
        return this.systemInfo ? this.systemInfo.with_clair : false;
    }

    public get isClairDBReady(): boolean {
        return (
            this.systemInfo &&
            this.systemInfo.clair_vulnerability_status &&
            this.systemInfo.clair_vulnerability_status.overall_last_update > 0
        );
    }

    public get withAdmiral(): boolean {
        return this.mode === "admiral";
    }

    public get showDBStatusWarning(): boolean {
        return this.withClair && !this.isClairDBReady;
    }

    ngOnChanges(changes: SimpleChanges): void {
        if (changes["projectId"] && changes["projectId"].currentValue) {
            this.refresh();
        }
    }

    ngOnInit(): void {
        // Get system info for tag views
        toPromise<SystemInfo>(this.systemInfoService.getSystemInfo())
            .then(systemInfo => (this.systemInfo = systemInfo))
            .catch(error => this.errorHandler.error(error));

        if (this.mode === "admiral") {
            this.isCardView = true;
        } else {
            this.isCardView = false;
        }

        this.lastFilteredRepoName = "";
    }

    confirmDeletion(message: ConfirmationAcknowledgement) {
        if (message &&
            message.source === ConfirmationTargets.REPOSITORY &&
            message.state === ConfirmationState.CONFIRMED) {

            let repoLists = message.data;
            if (repoLists && repoLists.length) {
                let promiseLists: any[] = [];
                repoLists.forEach(repo => {
                    promiseLists.push(this.delOperate(repo));
                });

                Promise.all(promiseLists).then((item) => {
                    this.selectedRow = [];
                    this.refresh();
                    let st: State = this.getStateAfterDeletion();
                    if (!st) {
                        this.refresh();
                    } else {
                        this.clrLoad(st);
                    }
                });
            }
        }
    }

    delOperate(repo: RepositoryItem) {
        // init operation info
        let operMessage = new OperateInfo();
        operMessage.name = 'OPERATION.DELETE_REPO';
        operMessage.data.id = repo.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = repo.name;
        this.operationService.publishInfo(operMessage);

        if (this.signedCon[repo.name].length !== 0) {
            forkJoin(this.translateService.get('BATCH.DELETED_FAILURE'),
                this.translateService.get('REPOSITORY.DELETION_TITLE_REPO_SIGNED')).subscribe(res => {
                operateChanges(operMessage, OperationState.failure, res[1]);
            });
        } else {
            return toPromise<number>(this.repositoryService
                .deleteRepository(repo.name))
                .then(
                    response => {
                        this.translateService.get('BATCH.DELETED_SUCCESS').subscribe(res => {
                            operateChanges(operMessage, OperationState.success);
                        });
                    }).catch(error => {
                    if (error.status === "412") {
                        forkJoin(this.translateService.get('BATCH.DELETED_FAILURE'),
                            this.translateService.get('REPOSITORY.TAGS_SIGNED')).subscribe(res => {
                            operateChanges(operMessage, OperationState.failure, res[1]);
                        });
                        return;
                    }
                    if (error.status === 503) {
                        forkJoin(this.translateService.get('BATCH.DELETED_FAILURE'),
                            this.translateService.get('REPOSITORY.TAGS_NO_DELETE')).subscribe(res => {
                            operateChanges(operMessage, OperationState.failure, res[1]);
                        });
                        return;
                    }
                    this.translateService.get('BATCH.DELETED_FAILURE').subscribe(res => {
                        operateChanges(operMessage, OperationState.failure, res);
                    });
                });
        }
    }

    doSearchRepoNames(repoName: string) {
        this.lastFilteredRepoName = repoName;
        this.currentPage = 1;
        let st: State = this.currentState;
        if (!st) {
            st = {page: {}};
        }
        st.page.size = this.pageSize;
        st.page.from = 0;
        st.page.to = this.pageSize - 1;
        this.clrLoad(st);
    }

    saveSignatures(event: { [key: string]: string[] }): void {
        Object.assign(this.signedCon, event);
    }

    deleteRepos(repoLists: RepositoryItem[]) {
        if (repoLists && repoLists.length) {
            let repoNames: string[] = [];
            let repArr: any[] = [];

            repoLists.forEach(repo => {
                repoNames.push(repo.name);

                if (!this.signedCon[repo.name]) {
                    repArr.push(this.getTagInfo(repo.name));
                }
            });

            Promise.all(repArr).then(() => {
                this.confirmationDialogSet(
                    'REPOSITORY.DELETION_TITLE_REPO',
                    '',
                    repoNames.join(','),
                    repoLists,
                    'REPOSITORY.DELETION_SUMMARY_REPO',
                    ConfirmationButtons.DELETE_CANCEL);
            });
        }
    }

    getTagInfo(repoName: string): Promise<void> {
        this.signedCon[repoName] = [];
        return toPromise<Tag[]>(this.tagService.getTags(repoName))
            .then(items => {
                items.forEach((t: Tag) => {
                    if (t.signature !== null) {
                        this.signedCon[repoName].push(t.name);
                    }
                });
            })
            .catch(error => this.errorHandler.error(error));
    }

    confirmationDialogSet(summaryTitle: string, signature: string,
        repoName: string, repoLists: RepositoryItem[],
        summaryKey: string, button: ConfirmationButtons): void {
        this.translateService.get(summaryKey,
            {
                repoName: repoName,
                signedImages: signature,
            }).pipe(finalize(() => {
                let hnd = setInterval(() => this.ref.markForCheck(), 100);
                setTimeout(() => clearInterval(hnd), 5000);
            }))
            .subscribe((res: string) => {
                summaryKey = res;
                let message = new ConfirmationMessage(
                    summaryTitle,
                    summaryKey,
                    repoName,
                    repoLists,
                    ConfirmationTargets.REPOSITORY,
                    button);
                this.confirmationDialog.open(message);


            });
    }

    containsLatestTag(repo: RepositoryItem): Promise<boolean> {
        return toPromise<Tag[]>(this.tagService.getTags(repo.name))
            .then(items => {
                if (items.some((t: Tag) => {
                        return t.name === 'latest';
                    })) {
                    return true;
                } else {
                    return false;
                }

            })
            .catch(error => Promise.reject(false));
    }

    provisionItemEvent(evt: any, repo: RepositoryItem): void {
        evt.stopPropagation();
        let repoCopy = clone(repo);
        repoCopy.name = this.registryUrl + ":443/" + repoCopy.name;
        this.containsLatestTag(repo)
            .then(containsLatest => {
                if (containsLatest) {
                    this.repoProvisionEvent.emit(repoCopy);
                } else {
                    this.addInfoEvent.emit(repoCopy);
                }
            })
            .catch(error => this.errorHandler.error(error));

    }

    itemAddInfoEvent(evt: any, repo: RepositoryItem): void {
        evt.stopPropagation();
        let repoCopy = clone(repo);
        repoCopy.name = this.registryUrl + ":443/" + repoCopy.name;
        this.addInfoEvent.emit(repoCopy);
    }

    deleteItemEvent(evt: any, item: RepositoryItem): void {
        evt.stopPropagation();
        this.deleteRepos([item]);
    }

    selectedChange(): void {
        let hnd = setInterval(() => this.ref.markForCheck(), 100);
        setTimeout(() => clearInterval(hnd), 2000);
    }

    refresh() {
        this.doSearchRepoNames("");
    }

    loadNextPage() {
        this.currentPage = this.currentPage + 1;
        // Pagination
        let params: RequestQueryParams = new RequestQueryParams();
        params.set("page", "" + this.currentPage);
        params.set("page_size", "" + this.pageSize);

        this.loading = true;
        toPromise<Repository>(
            this.repositoryService.getRepositories(
                this.projectId,
                this.lastFilteredRepoName,
                params
            )
        )
            .then((repo: Repository) => {
                this.totalCount = repo.metadata.xTotalCount;
                this.repositoriesCopy = repo.data;
                this.signedCon = {};
                // Do filtering and sorting
                this.repositoriesCopy = doFiltering<RepositoryItem>(
                    this.repositoriesCopy,
                    this.currentState
                );
                this.repositoriesCopy = doSorting<RepositoryItem>(
                    this.repositoriesCopy,
                    this.currentState
                );
                this.repositories = this.repositories.concat(this.repositoriesCopy);
                this.loading = false;
            })
            .catch(error => {
                this.loading = false;
                this.errorHandler.error(error);
            });
        let hnd = setInterval(() => this.ref.markForCheck(), 500);
        setTimeout(() => clearInterval(hnd), 5000);
    }

    clrLoad(state: State): void {
        this.selectedRow = [];
        // Keep it for future filtering and sorting
        this.currentState = state;

        let pageNumber: number = calculatePage(state);
        if (pageNumber <= 0) {
            pageNumber = 1;
        }

        // Pagination
        let params: RequestQueryParams = new RequestQueryParams();
        params.set("page", "" + pageNumber);
        params.set("page_size", "" + this.pageSize);

        this.loading = true;

        toPromise<Repository>(
            this.repositoryService.getRepositories(
                this.projectId,
                this.lastFilteredRepoName,
                params
            )
        )
            .then((repo: Repository) => {

                this.totalCount = repo.metadata.xTotalCount;
                this.repositories = repo.data;

                this.signedCon = {};
                // Do filtering and sorting
                this.repositories = doFiltering<RepositoryItem>(
                    this.repositories,
                    state
                );
                this.repositories = doSorting<RepositoryItem>(this.repositories, state);
                this.loading = false;
            })
            .catch(error => {
                this.loading = false;
                this.errorHandler.error(error);
            });

        // Force refresh view
        let hnd = setInterval(() => this.ref.markForCheck(), 100);
        setTimeout(() => clearInterval(hnd), 5000);
    }

    getStateAfterDeletion(): State {
        let total: number = this.totalCount - 1;
        if (total <= 0) {
            return null;
        }

        let totalPages: number = Math.ceil(total / this.pageSize);
        let targetPageNumber: number = this.currentPage;

        if (this.currentPage > totalPages) {
            targetPageNumber = totalPages; // Should == currentPage -1
        }

        let st: State = this.currentState;
        if (!st) {
            st = {page: {}};
        }
        st.page.size = this.pageSize;
        st.page.from = (targetPageNumber - 1) * this.pageSize;
        st.page.to = targetPageNumber * this.pageSize - 1;

        return st;
    }

    watchRepoClickEvt(repo: RepositoryItem) {
        this.repoClickEvent.emit(repo);
    }

    getImgLink(repo: RepositoryItem): string {
        return "/container-image-icons?container-image=" + repo.name;
    }

    showCard(cardView: boolean) {
        if (this.isCardView === cardView) {
            return;
        }
        this.isCardView = cardView;
        this.refresh();
    }

    mouseEnter(itemName: string) {
        if (itemName === "card") {
            this.cardHover = true;
        } else {
            this.listHover = true;
        }
    }

    mouseLeave(itemName: string) {
        if (itemName === "card") {
            this.cardHover = false;
        } else {
            this.listHover = false;
        }
    }

    isHovering(itemName: string) {
        if (itemName === "card") {
            return this.cardHover;
        } else {
            return this.listHover;
        }
    }
}
