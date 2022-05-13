import { Component, Input, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { forkJoin, Observable, of, Subject, Subscription } from 'rxjs';
import {
    catchError,
    debounceTime,
    distinctUntilChanged,
    finalize,
    map,
    switchMap,
} from 'rxjs/operators';
import { TranslateService } from '@ngx-translate/core';
import { NgForm } from '@angular/forms';
import { AVAILABLE_TIME } from '../artifact-list-page/artifact-list/artifact-list-tab/artifact-list-tab.component';
import { OperationService } from '../../../../../shared/components/operation/operation.service';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from '../../../../../shared/entities/shared.const';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../../shared/components/operation/operate';
import {
    AccessoryQueryParams,
    AccessoryType,
    ArtifactFront as Artifact,
    ArtifactType,
    getPullCommandByTag,
} from '../artifact';
import { ArtifactService } from '../../../../../../../ng-swagger-gen/services/artifact.service';
import { Tag } from '../../../../../../../ng-swagger-gen/models/tag';
import {
    SystemInfo,
    SystemInfoService,
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../../../shared/services';
import { ClrDatagridStateInterface } from '@clr/angular';
import {
    calculatePage,
    dbEncodeURIComponent,
    doFiltering,
    doSorting,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../../shared/units/utils';
import { AppConfigService } from '../../../../../services/app-config.service';
import { errorHandler } from '../../../../../shared/units/shared.utils';
import { ConfirmationDialogComponent } from '../../../../../shared/components/confirmation-dialog';
import { ConfirmationMessage } from '../../../../global-confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../../../../global-confirmation-dialog/confirmation-state-message';
import { ActivatedRoute } from '@angular/router';

class InitTag {
    name = '';
}
const DeleteTagWithNotoryCommand1 = 'notary -s https://';
const DeleteTagWithNotoryCommand2 = ':4443 -d ~/.docker/trust remove -p ';
@Component({
    selector: 'artifact-tag',
    templateUrl: './artifact-tag.component.html',
    styleUrls: ['./artifact-tag.component.scss'],
})
export class ArtifactTagComponent implements OnInit, OnDestroy {
    @Input() artifactDetails: Artifact;
    @Input() projectName: string;
    @Input() isProxyCacheProject: boolean = false;
    @Input() projectId: number;
    @Input() repositoryName: string;
    newTagName = new InitTag();
    newTagForm: NgForm;
    @ViewChild('newTagForm', { static: true }) currentForm: NgForm;
    selectedRow: Tag[] = [];
    isTagNameExist = false;
    newTagformShow = false;
    loading = true;
    openTag = false;
    availableTime = AVAILABLE_TIME;
    @ViewChild('confirmationDialog')
    confirmationDialog: ConfirmationDialogComponent;
    hasDeleteTagPermission: boolean;
    hasCreateTagPermission: boolean;

    totalCount: number = 0;
    allTags: Tag[] = [];
    currentTags: Tag[] = [];
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.ARTIFACT_TAGS_COMPONENT
    );
    currentPage = 1;
    tagNameChecker: Subject<string> = new Subject<string>();
    tagNameCheckSub: Subscription;
    tagNameCheckOnGoing = false;
    systemInfo: SystemInfo;
    accessoryType: string;
    constructor(
        private operationService: OperationService,
        private artifactService: ArtifactService,
        private translateService: TranslateService,
        private userPermissionService: UserPermissionService,
        private systemInfoService: SystemInfoService,
        private appConfigService: AppConfigService,
        private errorHandlerService: ErrorHandler,
        private activatedRoute: ActivatedRoute
    ) {
        this.accessoryType =
            this.activatedRoute.snapshot.queryParams[
                AccessoryQueryParams.ACCESSORY_TYPE
            ];
    }
    ngOnInit() {
        this.getImagePermissionRule(this.projectId);
        this.invalidCreateTag();
        this.systemInfoService.getSystemInfo().subscribe(
            systemInfo => (this.systemInfo = systemInfo),
            error => this.errorHandlerService.error(error)
        );
    }
    checkTagName(name) {
        const listTagParams: ArtifactService.ListTagsParams = {
            projectName: this.projectName,
            repositoryName: dbEncodeURIComponent(this.repositoryName),
            reference: this.artifactDetails.digest,
            withSignature: true,
            withImmutableStatus: true,
            q: encodeURIComponent(`name=${name}`),
        };
        return this.artifactService
            .listTags(listTagParams)
            .pipe(finalize(() => (this.tagNameCheckOnGoing = false)));
    }
    invalidCreateTag() {
        if (!this.tagNameCheckSub) {
            this.tagNameCheckSub = this.tagNameChecker
                .pipe(debounceTime(500))
                .pipe(distinctUntilChanged())
                .pipe(
                    switchMap(name => {
                        this.tagNameCheckOnGoing = true;
                        this.isTagNameExist = false;
                        return this.checkTagName(name);
                    })
                )
                .subscribe(
                    response => {
                        // tag existing
                        if (response && response.length) {
                            this.isTagNameExist = true;
                        }
                    },
                    error => {
                        this.errorHandlerService.error(error);
                    }
                );
        }
    }
    getCurrentArtifactTags(state: ClrDatagridStateInterface) {
        if (!state || !state.page) {
            return;
        }
        this.pageSize = state.page.size;
        setPageSizeToLocalStorage(
            PageSizeMapKeys.ARTIFACT_TAGS_COMPONENT,
            this.pageSize
        );
        let pageNumber: number = calculatePage(state);
        if (pageNumber <= 0) {
            pageNumber = 1;
        }
        let params: ArtifactService.ListTagsParams = {
            projectName: this.projectName,
            repositoryName: dbEncodeURIComponent(this.repositoryName),
            reference: this.artifactDetails.digest,
            page: pageNumber,
            withSignature: true,
            withImmutableStatus: true,
            pageSize: this.pageSize,
            sort: getSortingString(state),
        };
        this.loading = true;
        this.artifactService
            .listTagsResponse(params)
            .pipe(
                finalize(() => {
                    this.loading = false;
                })
            )
            .subscribe(
                res => {
                    if (res.headers) {
                        let xHeader: string = res.headers.get('x-total-count');
                        if (xHeader) {
                            this.totalCount = Number.parseInt(xHeader, 10);
                        }
                    }
                    this.currentTags = res.body;
                    // Do customising filtering and sorting
                    this.currentTags = doFiltering<Tag>(
                        this.currentTags,
                        state
                    );
                    this.currentTags = doSorting<Tag>(this.currentTags, state);
                },
                error => {
                    this.errorHandlerService.error(error);
                }
            );
    }
    getImagePermissionRule(projectId: number): void {
        const permissions = [
            {
                resource: USERSTATICPERMISSION.REPOSITORY_TAG.KEY,
                action: USERSTATICPERMISSION.REPOSITORY_TAG.VALUE.DELETE,
            },
            {
                resource: USERSTATICPERMISSION.REPOSITORY_TAG.KEY,
                action: USERSTATICPERMISSION.REPOSITORY_TAG.VALUE.CREATE,
            },
        ];
        this.userPermissionService
            .hasProjectPermissions(this.projectId, permissions)
            .subscribe(
                (results: Array<boolean>) => {
                    this.hasDeleteTagPermission = results[0];
                    this.hasCreateTagPermission = results[1];
                },
                error => this.errorHandlerService.error(error)
            );
    }

    addTag() {
        this.newTagformShow = true;
    }
    cancelAddTag() {
        this.newTagformShow = false;
        this.newTagName = new InitTag();
    }
    saveAddTag() {
        // const tag: NewTag = {name: this.newTagName};
        const createTagParams: ArtifactService.CreateTagParams = {
            projectName: this.projectName,
            repositoryName: dbEncodeURIComponent(this.repositoryName),
            reference: this.artifactDetails.digest,
            tag: this.newTagName,
        };
        this.loading = true;
        this.artifactService.createTag(createTagParams).subscribe(
            res => {
                this.newTagformShow = false;
                this.newTagName = new InitTag();
                this.refresh();
            },
            error => {
                this.loading = false;
                this.errorHandlerService.error(error);
            }
        );
    }
    removeTag() {
        if (this.selectedRow && this.selectedRow.length) {
            let tagNames: string[] = [];
            this.selectedRow.forEach(artifact => {
                tagNames.push(artifact.name);
            });
            let titleKey: string,
                summaryKey: string,
                content: string,
                buttons: ConfirmationButtons;
            titleKey = 'REPOSITORY.DELETION_TITLE_TAG';
            summaryKey = 'REPOSITORY.DELETION_SUMMARY_TAG';
            buttons = ConfirmationButtons.DELETE_CANCEL;
            content = tagNames.join(' , ');

            let message = new ConfirmationMessage(
                titleKey,
                summaryKey,
                content,
                this.selectedRow,
                ConfirmationTargets.TAG,
                buttons
            );
            this.confirmationDialog.open(message);
        }
    }
    confirmDeletion(message: ConfirmationAcknowledgement) {
        if (
            message &&
            message.source === ConfirmationTargets.TAG &&
            message.state === ConfirmationState.CONFIRMED
        ) {
            let tagList: Tag[] = message.data;
            if (tagList && tagList.length) {
                let observableLists: any[] = [];
                tagList.forEach(tag => {
                    observableLists.push(this.delOperate(tag));
                });
                this.loading = true;
                forkJoin(...observableLists).subscribe(deleteResult => {
                    // if delete one success  refresh list
                    let deleteSuccessList = [];
                    let deleteErrorList = [];
                    deleteResult.forEach(result => {
                        if (!result) {
                            // delete success
                            deleteSuccessList.push(result);
                        } else {
                            deleteErrorList.push(result);
                        }
                    });
                    this.selectedRow = [];
                    if (deleteSuccessList.length === deleteResult.length) {
                        // all is success
                        this.refresh();
                    } else if (deleteErrorList.length === deleteResult.length) {
                        // all is error
                        this.loading = false;
                        this.errorHandlerService.error(
                            deleteResult[deleteResult.length - 1]
                        );
                    } else {
                        // some artifact delete success but it has error delete things
                        this.errorHandlerService.error(
                            deleteErrorList[deleteErrorList.length - 1]
                        );
                        // if delete one success  refresh list
                        this.currentPage = 1;
                        let st: ClrDatagridStateInterface = {
                            page: {
                                from: 0,
                                to: this.pageSize - 1,
                                size: this.pageSize,
                            },
                        };
                        this.getCurrentArtifactTags(st);
                    }
                });
            }
        }
    }
    deletePort(url): string {
        if (url && url.indexOf(':') !== -1) {
            return url.split(':')[0];
        }
        return url;
    }
    delOperate(tag: Tag): Observable<any> | null {
        // init operation info
        let operMessage = new OperateInfo();
        operMessage.name = 'OPERATION.DELETE_TAG';
        operMessage.state = OperationState.progressing;
        operMessage.data.name = tag.name;
        this.operationService.publishInfo(operMessage);

        if (tag.signed) {
            forkJoin(
                this.translateService.get('BATCH.DELETED_FAILURE'),
                this.translateService.get(
                    'REPOSITORY.DELETION_SUMMARY_TAG_DENIED'
                )
            ).subscribe(res => {
                const wrongInfo: string =
                    res[1] +
                    DeleteTagWithNotoryCommand1 +
                    this.deletePort(this.registryUrl) +
                    DeleteTagWithNotoryCommand2 +
                    this.registryUrl +
                    '/' +
                    this.repositoryName +
                    ' ' +
                    tag.name;
                operateChanges(operMessage, OperationState.failure, wrongInfo);
            });
            return of(null);
        } else {
            const deleteTagParams: ArtifactService.DeleteTagParams = {
                projectName: this.projectName,
                repositoryName: dbEncodeURIComponent(this.repositoryName),
                reference: this.artifactDetails.digest,
                tagName: tag.name,
            };
            return this.artifactService.deleteTag(deleteTagParams).pipe(
                map(response => {
                    this.translateService
                        .get('BATCH.DELETED_SUCCESS')
                        .subscribe(res => {
                            operateChanges(operMessage, OperationState.success);
                        });
                }),
                catchError(error => {
                    const message = errorHandler(error);
                    this.translateService
                        .get(message)
                        .subscribe(res =>
                            operateChanges(
                                operMessage,
                                OperationState.failure,
                                res
                            )
                        );
                    return of(error);
                })
            );
        }
    }

    existValid(name) {
        if (name) {
            this.tagNameChecker.next(name);
        } else {
            this.isTagNameExist = false;
        }
    }
    toggleTagListOpenOrClose() {
        this.openTag = !this.openTag;
        this.newTagformShow = false;
    }
    hasImmutableOnTag(): boolean {
        return this.selectedRow.some(artifact => artifact.immutable);
    }
    refresh() {
        this.loading = true;
        this.currentPage = 1;
        this.selectedRow = [];
        let st: ClrDatagridStateInterface = {
            page: { from: 0, to: this.pageSize - 1, size: this.pageSize },
        };
        this.getCurrentArtifactTags(st);
    }
    ngOnDestroy(): void {
        this.tagNameCheckSub.unsubscribe();
    }
    get withNotary(): boolean {
        return this.appConfigService.getConfig().with_notary;
    }
    public get registryUrl(): string {
        if (this.systemInfo && this.systemInfo.registry_url) {
            return this.systemInfo.registry_url;
        }
        return location.hostname;
    }
    hasPullCommand(): boolean {
        return (
            this.artifactDetails &&
            (this.artifactDetails.type === ArtifactType.IMAGE ||
                this.artifactDetails.type === ArtifactType.CHART ||
                this.artifactDetails.type === ArtifactType.CNAB) &&
            this.accessoryType !== AccessoryType.COSIGN
        );
    }
    getPullCommand(tag: Tag): string {
        let pullCommand: string = '';
        if (tag && tag.name && this.artifactDetails) {
            pullCommand = getPullCommandByTag(
                this.artifactDetails?.type,
                `${this.registryUrl}/${this.projectName}/${this.repositoryName}`,
                tag.name
            );
        }
        return pullCommand;
    }
}
