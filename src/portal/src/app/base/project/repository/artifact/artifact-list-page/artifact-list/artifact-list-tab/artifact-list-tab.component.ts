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
    ElementRef,
    OnDestroy,
    OnInit,
    ViewChild,
} from '@angular/core';
import { forkJoin, Observable, of, Subject, Subscription } from 'rxjs';
import {
    catchError,
    debounceTime,
    distinctUntilChanged,
    finalize,
    map,
} from 'rxjs/operators';
import { TranslateService } from '@ngx-translate/core';
import {
    ClrDatagridComparatorInterface,
    ClrDatagridStateInterface,
    ClrLoadingState,
} from '@clr/angular';
import { ActivatedRoute, Router } from '@angular/router';
import { Comparator } from '../../../../../../../shared/services';
import {
    calculatePage,
    clone,
    CustomComparator,
    dbEncodeURIComponent,
    DEFAULT_SUPPORTED_MIME_TYPES,
    doSorting,
    formatSize,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
    VULNERABILITY_SCAN_STATUS,
} from '../../../../../../../shared/units/utils';
import { ImageNameInputComponent } from '../../../../../../../shared/components/image-name-input/image-name-input.component';
import { ErrorHandler } from '../../../../../../../shared/units/error-handler';
import { ArtifactService } from '../../../artifact.service';
import { OperationService } from '../../../../../../../shared/components/operation/operation.service';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from '../../../../../../../shared/entities/shared.const';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../../../../shared/components/operation/operate';
import {
    AccessoryType,
    artifactDefault,
    ArtifactFront as Artifact,
    ArtifactFront,
    ArtifactType,
    getPullCommandByDigest,
    getPullCommandByTag,
    mutipleFilter,
} from '../../../artifact';
import { Project } from '../../../../../project';
import { ArtifactService as NewArtifactService } from '../../../../../../../../../ng-swagger-gen/services/artifact.service';
import { ADDITIONS } from '../../../artifact-additions/models';
import { Platform } from '../../../../../../../../../ng-swagger-gen/models/platform';
import { SafeUrl } from '@angular/platform-browser';
import { errorHandler } from '../../../../../../../shared/units/shared.utils';
import { ConfirmationDialogComponent } from '../../../../../../../shared/components/confirmation-dialog';
import { ConfirmationMessage } from '../../../../../../global-confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../../../../../../global-confirmation-dialog/confirmation-state-message';
import {
    UN_LOGGED_PARAM,
    YES,
} from '../../../../../../../account/sign-in/sign-in.service';
import { Label } from '../../../../../../../../../ng-swagger-gen/models/label';
import {
    EventService,
    HarborEvent,
} from '../../../../../../../services/event-service/event.service';
import { AppConfigService } from 'src/app/services/app-config.service';
import { ArtifactListPageService } from '../../artifact-list-page.service';
import { ACCESSORY_PAGE_SIZE } from './sub-accessories/sub-accessories.component';
import { Accessory } from 'ng-swagger-gen/models/accessory';
import { Tag } from '../../../../../../../../../ng-swagger-gen/models/tag';

export interface LabelState {
    iconsShow: boolean;
    label: Label;
    show: boolean;
}

export const AVAILABLE_TIME = '0001-01-01T00:00:00.000Z';

const CHECKING: string = 'checking';
const TRUE: string = 'true';
const FALSE: string = 'false';

@Component({
    selector: 'artifact-list-tab',
    templateUrl: './artifact-list-tab.component.html',
    styleUrls: ['./artifact-list-tab.component.scss'],
})
export class ArtifactListTabComponent implements OnInit, OnDestroy {
    projectId: number;
    projectName: string;
    repoName: string;
    registryUrl: string;
    artifactList: ArtifactFront[] = [];
    availableTime = AVAILABLE_TIME;
    showTagManifestOpened: boolean;
    retagDialogOpened: boolean;
    manifestInfoTitle: string;
    digestId: string;
    staticBackdrop = true;
    closable = false;
    lastFilteredTagName: string;
    inprogress: boolean;
    openLabelFilterPanel: boolean;
    openLabelFilterPiece: boolean;
    retagSrcImage: string;
    showlabel: boolean;

    pullComparator: Comparator<Artifact> = new CustomComparator<Artifact>(
        'pull_time',
        'date'
    );
    pushComparator: Comparator<Artifact> = new CustomComparator<Artifact>(
        'push_time',
        'date'
    );

    loading = true;
    copyFailed = false;
    selectedRow: Artifact[] = [];
    labelListOpen = false;
    selectedTag: Artifact[];
    labelNameFilter: Subject<string> = new Subject<string>();
    stickLabelNameFilter: Subject<string> = new Subject<string>();
    filterOnGoing: boolean;
    stickName = '';
    filterName = '';
    initFilter = {
        name: '',
        description: '',
        color: '',
        scope: '',
        project_id: 0,
    };
    filterOneLabel: Label = this.initFilter;

    @ViewChild('confirmationDialog')
    confirmationDialog: ConfirmationDialogComponent;

    @ViewChild('imageNameInput')
    imageNameInput: ImageNameInputComponent;

    @ViewChild('digestTarget') textInput: ElementRef;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.ARTIFACT_LIST_TAB_COMPONENT
    );
    currentPage = 1;
    totalCount = 0;
    currentState: ClrDatagridStateInterface;

    get hasAddLabelImagePermission(): boolean {
        return this.artifactListPageService.hasAddLabelImagePermission();
    }
    get hasRetagImagePermission(): boolean {
        return this.artifactListPageService.hasRetagImagePermission();
    }
    get hasDeleteImagePermission(): boolean {
        return this.artifactListPageService.hasDeleteImagePermission();
    }
    get hasScanImagePermission(): boolean {
        return this.artifactListPageService.hasScanImagePermission();
    }
    get hasEnabledScanner(): boolean {
        return this.artifactListPageService.hasEnabledScanner();
    }
    get scanBtnState(): ClrLoadingState {
        return this.artifactListPageService.getScanBtnState();
    }
    onSendingScanCommand: boolean;
    onSendingStopScanCommand: boolean = false;
    onStopScanArtifactsLength: number = 0;
    scanStoppedArtifactLength: number = 0;

    artifactDigest: string;
    depth: string;
    hasInit: boolean = false;
    triggerSub: Subscription;
    labelNameFilterSub: Subscription;
    stickLabelNameFilterSub: Subscription;
    mutipleFilter = clone(mutipleFilter);
    filterByType: string = this.mutipleFilter[0].filterBy;
    openSelectFilterPiece = false;
    // could Pagination filter
    filters: string[];

    scanFinishedArtifactLength: number = 0;
    onScanArtifactsLength: number = 0;
    stopBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    updateArtifactSub: Subscription;
    constructor(
        private errorHandlerService: ErrorHandler,
        private artifactService: ArtifactService,
        private newArtifactService: NewArtifactService,
        private translateService: TranslateService,
        private operationService: OperationService,
        private eventService: EventService,
        private activatedRoute: ActivatedRoute,
        private router: Router,
        private appConfigService: AppConfigService,
        public artifactListPageService: ArtifactListPageService
    ) {}

    ngOnInit() {
        this.artifactListPageService.resetClonedLabels();
        this.projectId =
            this.activatedRoute.snapshot?.parent?.parent?.params['id'];
        let resolverData = this.activatedRoute.snapshot?.parent?.parent?.data;
        if (resolverData) {
            this.projectName = (<Project>resolverData['projectResolver']).name;
        }
        this.repoName = this.activatedRoute.snapshot?.parent?.params['repo'];
        this.registryUrl = this.appConfigService.getConfig().registry_url;
        this.activatedRoute.params?.subscribe(params => {
            this.depth =
                this.activatedRoute.snapshot?.firstChild?.params['depth'];
            if (this.depth) {
                const arr: string[] = this.depth.split('-');
                this.artifactDigest = this.depth.split('-')[arr.length - 1];
            }
            if (this.hasInit) {
                this.currentPage = 1;
                this.totalCount = 0;
                const st: ClrDatagridStateInterface = {
                    page: {
                        from: 0,
                        to: this.pageSize - 1,
                        size: this.pageSize,
                    },
                };
                this.clrLoad(st);
            }
            this.init();
        });
        if (!this.updateArtifactSub) {
            this.updateArtifactSub = this.eventService.subscribe(
                HarborEvent.UPDATE_VULNERABILITY_INFO,
                (artifact: Artifact) => {
                    if (this.artifactList && this.artifactList.length) {
                        this.artifactList.forEach(item => {
                            if (item.digest === artifact.digest) {
                                item.scan_overview = artifact.scan_overview;
                            }
                        });
                    }
                }
            );
        }
    }

    ngOnDestroy() {
        if (this.triggerSub) {
            this.triggerSub.unsubscribe();
            this.triggerSub = null;
        }
        if (this.labelNameFilterSub) {
            this.labelNameFilterSub.unsubscribe();
            this.labelNameFilterSub = null;
        }
        if (this.stickLabelNameFilterSub) {
            this.stickLabelNameFilterSub.unsubscribe();
            this.stickLabelNameFilterSub = null;
        }
        if (this.updateArtifactSub) {
            this.updateArtifactSub.unsubscribe();
            this.updateArtifactSub = null;
        }
    }

    init() {
        this.hasInit = true;
        this.depth = this.activatedRoute.snapshot.params['depth'];
        if (this.depth) {
            const arr: string[] = this.depth.split('-');
            this.artifactDigest = this.depth.split('-')[arr.length - 1];
        }
        if (!this.projectId) {
            this.errorHandlerService.error('Project ID cannot be unset.');
            return;
        }
        const resolverData = this.activatedRoute.snapshot.params.data;
        if (resolverData) {
            const pro: Project = <Project>resolverData['projectResolver'];
            this.projectName = pro.name;
        }
        if (!this.repoName) {
            this.errorHandlerService.error('Repo name cannot be unset.');
            return;
        }
        this.lastFilteredTagName = '';
        if (!this.labelNameFilterSub) {
            this.labelNameFilterSub = this.labelNameFilter
                .pipe(debounceTime(500))
                .pipe(distinctUntilChanged())
                .subscribe((name: string) => {
                    if (this.filterName.length) {
                        this.filterOnGoing = true;
                        this.artifactListPageService.imageFilterLabels.forEach(
                            data => {
                                data.show =
                                    data.label.name.indexOf(this.filterName) !==
                                    -1;
                            }
                        );
                    }
                });
        }
        if (!this.stickLabelNameFilterSub) {
            this.stickLabelNameFilterSub = this.stickLabelNameFilter
                .pipe(debounceTime(500))
                .pipe(distinctUntilChanged())
                .subscribe((name: string) => {
                    if (this.stickName.length) {
                        this.filterOnGoing = true;
                        this.artifactListPageService.imageStickLabels.forEach(
                            data => {
                                data.show =
                                    data.label.name.indexOf(this.stickName) !==
                                    -1;
                            }
                        );
                    }
                });
        }
    }

    public get filterLabelPieceWidth() {
        let len = this.lastFilteredTagName.length
            ? this.lastFilteredTagName.length * 6 + 60
            : 115;
        return len > 210 ? 210 : len;
    }

    get withNotary(): boolean {
        return this.appConfigService.getConfig()?.with_notary;
    }

    doSearchArtifactByFilter(filterWords) {
        this.lastFilteredTagName = filterWords;
        this.currentPage = 1;

        let st: ClrDatagridStateInterface = this.currentState;
        if (!st) {
            st = { page: {} };
        }
        st.page.size = this.pageSize;
        st.page.from = 0;
        st.page.to = this.pageSize - 1;
        this.filters = [];
        if (this.lastFilteredTagName) {
            this.filters.push(
                `${this.filterByType}=~${this.lastFilteredTagName}`
            );
        }
        this.clrLoad(st);
    }

    // todo
    clrDgRefresh(state: ClrDatagridStateInterface) {
        setTimeout(() => {
            this.clrLoad(state);
        });
    }

    clrLoad(state: ClrDatagridStateInterface): void {
        this.artifactList = [];
        this.loading = true;
        if (!state || !state.page) {
            return;
        }
        this.pageSize = state.page.size;
        setPageSizeToLocalStorage(
            PageSizeMapKeys.ARTIFACT_LIST_TAB_COMPONENT,
            this.pageSize
        );
        this.selectedRow = [];
        // Keep it for future filtering and sorting

        let pageNumber: number = calculatePage(state);
        if (pageNumber <= 0) {
            pageNumber = 1;
        }
        let sortBy: any = '';
        if (state.sort) {
            sortBy = state.sort.by as
                | string
                | ClrDatagridComparatorInterface<any>;
            sortBy = sortBy.fieldName ? sortBy.fieldName : sortBy;
            sortBy = state.sort.reverse ? `-${sortBy}` : sortBy;
        }
        this.currentState = state;

        // Pagination
        let params: any = {};
        if (pageNumber && this.pageSize) {
            params.page = pageNumber;
            params.pageSize = this.pageSize;
        }
        if (sortBy) {
            params.sort = sortBy;
        }
        if (this.filters && this.filters.length) {
            let q = '';
            this.filters.forEach(item => {
                q += item;
            });
            params.q = encodeURIComponent(q);
        }
        if (this.artifactDigest) {
            const artifactParam: NewArtifactService.GetArtifactParams = {
                repositoryName: dbEncodeURIComponent(this.repoName),
                projectName: this.projectName,
                reference: this.artifactDigest,
                withImmutableStatus: true,
                withLabel: true,
                withScanOverview: true,
                withTag: false,
                XAcceptVulnerabilities: DEFAULT_SUPPORTED_MIME_TYPES,
                withAccessory: false,
            };
            this.newArtifactService.getArtifact(artifactParam).subscribe(
                res => {
                    let observableLists: Observable<Artifact>[] = [];
                    let platFormAttr: { platform: Platform }[] = [];
                    this.totalCount = res.references.length;
                    res.references.forEach((child, index) => {
                        if (
                            index >= (pageNumber - 1) * this.pageSize &&
                            index < pageNumber * this.pageSize
                        ) {
                            let childParams: NewArtifactService.GetArtifactParams =
                                {
                                    repositoryName: dbEncodeURIComponent(
                                        this.repoName
                                    ),
                                    projectName: this.projectName,
                                    reference: child.child_digest,
                                    withImmutableStatus: true,
                                    withLabel: true,
                                    withScanOverview: true,
                                    withTag: false,
                                    XAcceptVulnerabilities:
                                        DEFAULT_SUPPORTED_MIME_TYPES,
                                    withAccessory: false,
                                };
                            platFormAttr.push({ platform: child.platform });
                            observableLists.push(
                                this.newArtifactService.getArtifact(childParams)
                            );
                        }
                    });
                    forkJoin(observableLists)
                        .pipe(
                            finalize(() => {
                                this.loading = false;
                            })
                        )
                        .subscribe(
                            artifacts => {
                                this.artifactList = artifacts;
                                this.artifactList = doSorting<ArtifactFront>(
                                    this.artifactList,
                                    state
                                );
                                this.artifactList.forEach((artifact, index) => {
                                    artifact.platform = clone(
                                        platFormAttr[index].platform
                                    );
                                });
                                this.getArtifactTagsAsync(this.artifactList);
                                this.getAccessoriesAsync(this.artifactList);
                                this.checkCosignAsync(this.artifactList);
                                this.getIconsFromBackEnd();
                            },
                            error => {
                                this.errorHandlerService.error(error);
                            }
                        );
                },
                error => {
                    this.loading = false;
                }
            );
        } else {
            let listArtifactParams: NewArtifactService.ListArtifactsParams = {
                projectName: this.projectName,
                repositoryName: dbEncodeURIComponent(this.repoName),
                withLabel: true,
                withScanOverview: true,
                withTag: false,
                sort: getSortingString(state),
                XAcceptVulnerabilities: DEFAULT_SUPPORTED_MIME_TYPES,
                withAccessory: false,
            };
            Object.assign(listArtifactParams, params);
            this.newArtifactService
                .listArtifactsResponse(listArtifactParams)
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    res => {
                        if (res.headers) {
                            let xHeader: string =
                                res.headers.get('X-Total-Count');
                            if (xHeader) {
                                this.totalCount = parseInt(xHeader, 0);
                            }
                        }
                        this.artifactList = res.body;
                        this.getArtifactTagsAsync(this.artifactList);
                        this.getAccessoriesAsync(this.artifactList);
                        this.checkCosignAsync(this.artifactList);
                        this.getIconsFromBackEnd();
                    },
                    error => {
                        // error
                        this.errorHandlerService.error(error);
                    }
                );
        }
    }

    refresh() {
        this.currentPage = 1;
        let st: ClrDatagridStateInterface = this.currentState;
        if (!st) {
            st = { page: {} };
            st.page.size = this.pageSize;
            st.page.from = 0;
            st.page.to = this.pageSize - 1;
        }
        this.clrLoad(st);
    }

    getPullCommand(artifact: Artifact): string {
        let pullCommand: string = '';
        if (
            artifact.type === ArtifactType.CHART &&
            artifact.tags &&
            artifact.tags[0]
        ) {
            pullCommand = getPullCommandByTag(
                artifact.type,
                `${this.registryUrl ? this.registryUrl : location.hostname}/${
                    this.projectName
                }/${this.repoName}`,
                artifact.tags[0]?.name
            );
        } else {
            pullCommand = getPullCommandByDigest(
                artifact.type,
                `${this.registryUrl ? this.registryUrl : location.hostname}/${
                    this.projectName
                }/${this.repoName}`,
                artifact.digest
            );
        }
        return pullCommand;
    }
    labelSelectedChange(artifact?: Artifact[]): void {
        this.artifactListPageService.imageStickLabels.forEach(data => {
            data.iconsShow = false;
            data.show = true;
        });
        if (artifact && artifact[0].labels && artifact[0].labels.length) {
            artifact[0].labels.forEach((labelInfo: Label) => {
                let findedLabel =
                    this.artifactListPageService.imageStickLabels.find(
                        data => labelInfo.id === data['label'].id
                    );
                if (findedLabel) {
                    this.artifactListPageService.imageStickLabels.splice(
                        this.artifactListPageService.imageStickLabels.indexOf(
                            findedLabel
                        ),
                        1
                    );
                    this.artifactListPageService.imageStickLabels.unshift(
                        findedLabel
                    );
                    findedLabel.iconsShow = true;
                }
            });
        }
    }

    addLabels(): void {
        this.labelListOpen = true;
        this.selectedTag = this.selectedRow;
        this.stickName = '';
        this.labelSelectedChange(this.selectedRow);
    }

    canAddLabel(): boolean {
        if (this.selectedRow && this.selectedRow.length === 1) {
            return true;
        }
        if (this.selectedRow && this.selectedRow.length > 1) {
            for (let i = 0; i < this.selectedRow.length; i++) {
                if (
                    this.selectedRow[i].labels &&
                    this.selectedRow[i].labels.length
                ) {
                    return false;
                }
            }
            return true;
        }
        return false;
    }

    stickLabel(labelInfo: LabelState): void {
        if (labelInfo && !labelInfo.iconsShow) {
            this.selectLabel(labelInfo);
        }
        if (labelInfo && labelInfo.iconsShow) {
            this.unSelectLabel(labelInfo);
        }
    }

    selectLabel(labelInfo: LabelState): void {
        if (!this.inprogress) {
            // add label to multiple artifact
            const ObservableArr: Array<Observable<null>> = [];
            this.selectedRow.forEach(item => {
                const params: NewArtifactService.AddLabelParams = {
                    projectName: this.projectName,
                    repositoryName: dbEncodeURIComponent(this.repoName),
                    reference: item.digest,
                    label: labelInfo.label,
                };
                ObservableArr.push(this.newArtifactService.addLabel(params));
            });
            this.inprogress = true;
            forkJoin(ObservableArr)
                .pipe(finalize(() => (this.inprogress = false)))
                .subscribe(
                    res => {
                        this.refresh();
                        // set the selected label in front
                        this.artifactListPageService.imageStickLabels.splice(
                            this.artifactListPageService.imageStickLabels.indexOf(
                                labelInfo
                            ),
                            1
                        );
                        this.artifactListPageService.imageStickLabels.some(
                            (data, i) => {
                                if (!data.iconsShow) {
                                    this.artifactListPageService.imageStickLabels.splice(
                                        i,
                                        0,
                                        labelInfo
                                    );
                                    return true;
                                }
                            }
                        );
                        // when is the last one
                        if (
                            this.artifactListPageService.imageStickLabels.every(
                                data => data.iconsShow === true
                            )
                        ) {
                            this.artifactListPageService.imageStickLabels.push(
                                labelInfo
                            );
                        }
                        labelInfo.iconsShow = true;
                    },
                    err => {
                        this.refresh();
                        this.errorHandlerService.error(err);
                    }
                );
        }
    }

    unSelectLabel(labelInfo: LabelState): void {
        if (!this.inprogress) {
            this.inprogress = true;
            let labelId = labelInfo.label.id;
            this.selectedRow = this.selectedTag;
            let params: NewArtifactService.RemoveLabelParams = {
                projectName: this.projectName,
                repositoryName: dbEncodeURIComponent(this.repoName),
                reference: this.selectedRow[0].digest,
                labelId: labelId,
            };
            this.newArtifactService.removeLabel(params).subscribe(
                res => {
                    this.refresh();

                    // insert the unselected label to groups with the same icons
                    this.sortOperation(
                        this.artifactListPageService.imageStickLabels,
                        labelInfo
                    );
                    labelInfo.iconsShow = false;
                    this.inprogress = false;
                },
                err => {
                    this.inprogress = false;
                    this.errorHandlerService.error(err);
                }
            );
        }
    }

    rightFilterLabel(labelInfo: LabelState): void {
        if (labelInfo) {
            if (!labelInfo.iconsShow) {
                this.filterLabel(labelInfo);
                this.showlabel = true;
            } else {
                this.unFilterLabel(labelInfo);
                this.showlabel = false;
            }
        }
    }

    filterLabel(labelInfo: LabelState): void {
        let labelId = labelInfo.label.id;
        this.artifactListPageService.imageFilterLabels.filter(data => {
            data.iconsShow = data.label.id === labelId;
        });
        this.filterOneLabel = labelInfo.label;

        // reload data
        this.currentPage = 1;
        let st: ClrDatagridStateInterface = this.currentState;
        if (!st) {
            st = { page: {} };
        }
        st.page.size = this.pageSize;
        st.page.from = 0;
        st.page.to = this.pageSize - 1;

        this.filters = [`${this.filterByType}=(${labelId})`];

        this.clrLoad(st);
    }

    unFilterLabel(labelInfo: LabelState): void {
        this.filterOneLabel = this.initFilter;
        labelInfo.iconsShow = false;
        // reload data
        this.currentPage = 1;
        let st: ClrDatagridStateInterface = this.currentState;
        if (!st) {
            st = { page: {} };
        }
        st.page.size = this.pageSize;
        st.page.from = 0;
        st.page.to = this.pageSize - 1;

        this.filters = [];
        this.clrLoad(st);
    }

    closeFilter(): void {
        this.openLabelFilterPanel = false;
    }

    reSortImageFilterLabels() {
        if (
            this.artifactListPageService.imageFilterLabels &&
            this.artifactListPageService.imageFilterLabels.length
        ) {
            for (
                let i = 0;
                i < this.artifactListPageService.imageFilterLabels.length;
                i++
            ) {
                if (
                    this.artifactListPageService.imageFilterLabels[i].iconsShow
                ) {
                    const arr: LabelState[] =
                        this.artifactListPageService.imageFilterLabels.splice(
                            i,
                            1
                        );
                    this.artifactListPageService.imageFilterLabels.unshift(
                        ...arr
                    );
                    break;
                }
            }
        }
    }

    getFilterPlaceholder(): string {
        return this.showlabel ? '' : 'ARTIFACT.FILTER_FOR_ARTIFACTS';
    }

    openFlagEvent(isOpen: boolean): void {
        if (isOpen) {
            this.openLabelFilterPanel = true;
            // every time  when filer panel opens, resort imageFilterLabels labels
            this.reSortImageFilterLabels();
            this.openLabelFilterPiece = true;
            this.openSelectFilterPiece = true;
            this.filterName = '';
            // redisplay all labels
            this.artifactListPageService.imageFilterLabels.forEach(data => {
                data.show = data.label.name.indexOf(this.filterName) !== -1;
            });
        } else {
            this.openLabelFilterPanel = false;
            this.openLabelFilterPiece = false;
            this.openSelectFilterPiece = false;
        }
    }

    handleInputFilter() {
        if (this.filterName.length) {
            this.labelNameFilter.next(this.filterName);
        } else {
            this.artifactListPageService.imageFilterLabels.every(
                data => (data.show = true)
            );
        }
    }

    handleStickInputFilter() {
        if (this.stickName.length) {
            this.stickLabelNameFilter.next(this.stickName);
        } else {
            this.artifactListPageService.imageStickLabels.every(
                data => (data.show = true)
            );
        }
    }

    // insert the unselected label to groups with the same icons
    sortOperation(labelList: LabelState[], labelInfo: LabelState): void {
        labelList.some((data, i) => {
            if (!data.iconsShow) {
                if (data.label.scope === labelInfo.label.scope) {
                    labelList.splice(i, 0, labelInfo);
                    labelList.splice(labelList.indexOf(labelInfo, 0), 1);
                    return true;
                }
                if (
                    data.label.scope !== labelInfo.label.scope &&
                    i === labelList.length - 1
                ) {
                    labelList.push(labelInfo);
                    labelList.splice(labelList.indexOf(labelInfo), 1);
                    return true;
                }
            }
        });
    }

    sizeTransform(tagSize: string): string {
        return formatSize(tagSize);
    }

    retag() {
        if (this.selectedRow && this.selectedRow.length && !this.depth) {
            this.retagDialogOpened = true;
            this.imageNameInput.imageNameForm.reset({
                repoName: this.repoName,
            });
            this.retagSrcImage =
                this.repoName + ':' + this.selectedRow[0].digest;
        }
    }

    onRetag() {
        let params: NewArtifactService.CopyArtifactParams = {
            projectName: this.imageNameInput.projectName.value,
            repositoryName: dbEncodeURIComponent(
                this.imageNameInput.repoName.value
            ),
            from: `${this.projectName}/${this.repoName}@${this.selectedRow[0].digest}`,
        };
        this.newArtifactService
            .CopyArtifact(params)
            .pipe(
                finalize(() => {
                    this.imageNameInput.form.reset();
                    this.retagDialogOpened = false;
                })
            )
            .subscribe(
                response => {
                    this.translateService
                        .get('RETAG.MSG_SUCCESS')
                        .subscribe((res: string) => {
                            this.errorHandlerService.info(res);
                        });
                },
                error => {
                    this.errorHandlerService.error(error);
                }
            );
    }

    deleteArtifact() {
        if (this.selectedRow && this.selectedRow.length && !this.depth) {
            let artifactNames: string[] = [];
            this.selectedRow.forEach(artifact => {
                artifactNames.push(artifact.digest.slice(0, 15));
            });

            let titleKey: string,
                summaryKey: string,
                content: string,
                buttons: ConfirmationButtons;
            titleKey = 'REPOSITORY.DELETION_TITLE_ARTIFACT';
            summaryKey = 'REPOSITORY.DELETION_SUMMARY_ARTIFACT';
            buttons = ConfirmationButtons.DELETE_CANCEL;
            content = artifactNames.join(' , ');
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

    deleteArtifactobservableLists: Observable<any>[] = [];

    confirmDeletion(message: ConfirmationAcknowledgement) {
        if (
            message &&
            message.source === ConfirmationTargets.ACCESSORY &&
            message.state === ConfirmationState.CONFIRMED
        ) {
            // delete one  accessory
            this.loading = true;
            // init operation info
            const opeMessage = new OperateInfo();
            opeMessage.name = 'ACCESSORY.DELETE_ACCESSORY';
            opeMessage.data.id = message.data.id;
            opeMessage.state = OperationState.progressing;
            opeMessage.data.name = message.data.digest.slice(0, 15);
            this.operationService.publishInfo(opeMessage);
            const params: NewArtifactService.DeleteArtifactParams = {
                projectName: this.projectName,
                repositoryName: dbEncodeURIComponent(this.repoName),
                reference: message.data.digest,
            };
            this.newArtifactService
                .deleteArtifact(params)
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    res => {
                        this.errorHandlerService.info(
                            'ACCESSORY.DELETED_SUCCESS'
                        );
                        operateChanges(opeMessage, OperationState.success);
                        this.refresh();
                    },
                    error => {
                        this.errorHandlerService.error(error);
                        operateChanges(
                            opeMessage,
                            OperationState.failure,
                            errorHandler(error)
                        );
                    }
                );
        }
        if (
            message &&
            message.source === ConfirmationTargets.TAG &&
            message.state === ConfirmationState.CONFIRMED
        ) {
            let artifactList = message.data;
            if (artifactList && artifactList.length) {
                artifactList.forEach(artifact => {
                    this.deleteArtifactobservableLists.push(
                        this.delOperate(artifact)
                    );
                });
                this.loading = true;
                forkJoin(...this.deleteArtifactobservableLists).subscribe(
                    deleteResult => {
                        let deleteSuccessList = [];
                        let deleteErrorList = [];
                        this.deleteArtifactobservableLists = [];
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
                            let st: ClrDatagridStateInterface = {
                                page: {
                                    from: 0,
                                    to: this.pageSize - 1,
                                    size: this.pageSize,
                                },
                            };
                            this.clrLoad(st);
                        } else if (
                            deleteErrorList.length === deleteResult.length
                        ) {
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
                            let st: ClrDatagridStateInterface = {
                                page: {
                                    from: 0,
                                    to: this.pageSize - 1,
                                    size: this.pageSize,
                                },
                            };
                            this.clrLoad(st);
                        }
                    }
                );
            }
        }
    }

    delOperate(artifact: Artifact): Observable<any> | null {
        // init operation info
        let operMessage = new OperateInfo();
        operMessage.name = 'OPERATION.DELETE_TAG';
        operMessage.data.id = artifact.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = artifact.digest;
        this.operationService.publishInfo(operMessage);
        let params: NewArtifactService.DeleteArtifactParams = {
            projectName: this.projectName,
            repositoryName: dbEncodeURIComponent(this.repoName),
            reference: artifact.digest,
        };
        return this.newArtifactService.deleteArtifact(params).pipe(
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
                        operateChanges(operMessage, OperationState.failure, res)
                    );
                return of(error);
            })
        );
        // }
    }

    showDigestId() {
        if (this.selectedRow && this.selectedRow.length === 1 && !this.depth) {
            this.manifestInfoTitle = 'REPOSITORY.COPY_DIGEST_ID';
            this.digestId = this.selectedRow[0].digest;
            this.showTagManifestOpened = true;
            this.copyFailed = false;
        }
    }

    goIntoArtifactSummaryPage(artifact: Artifact): void {
        const relativeRouterLink: string[] = ['artifacts', artifact.digest];
        if (this.activatedRoute.snapshot.queryParams[UN_LOGGED_PARAM] === YES) {
            this.router.navigate(relativeRouterLink, {
                relativeTo: this.activatedRoute,
                queryParams: { [UN_LOGGED_PARAM]: YES },
            });
        } else {
            this.router.navigate(relativeRouterLink, {
                relativeTo: this.activatedRoute,
            });
        }
    }

    onSuccess($event: any): void {
        this.copyFailed = false;
        // Directly close dialog
        this.showTagManifestOpened = false;
    }

    onError($event: any): void {
        // Show error
        this.copyFailed = true;
        // Select all text
        if (this.textInput) {
            this.textInput.nativeElement.select();
        }
    }

    // Get vulnerability scanning status
    scanStatus(artifact: Artifact): string {
        if (artifact) {
            let so = this.handleScanOverview((<any>artifact).scan_overview);
            if (so && so.scan_status) {
                return so.scan_status;
            }
        }
        return VULNERABILITY_SCAN_STATUS.NOT_SCANNED;
    }

    // if has running job, return false
    canScanNow(): boolean {
        if (!this.hasScanImagePermission) {
            return false;
        }
        if (this.onSendingScanCommand) {
            return false;
        }
        if (this.selectedRow && this.selectedRow.length) {
            let flag: boolean = true;
            this.selectedRow.forEach(item => {
                const st: string = this.scanStatus(item);
                if (this.isRunningState(st)) {
                    flag = false;
                }
            });
            return flag;
        }
        return false;
    }
    // Trigger scan
    scanNow(): void {
        if (!this.selectedRow.length) {
            return;
        }
        this.scanFinishedArtifactLength = 0;
        this.onScanArtifactsLength = this.selectedRow.length;
        this.onSendingScanCommand = true;
        this.selectedRow.forEach((data: any) => {
            let digest = data.digest;
            this.eventService.publish(
                HarborEvent.START_SCAN_ARTIFACT,
                this.repoName + '/' + digest
            );
        });
    }

    selectedRowHasVul(): boolean {
        return !!(
            this.selectedRow &&
            this.selectedRow[0] &&
            this.selectedRow[0].addition_links &&
            this.selectedRow[0].addition_links[ADDITIONS.VULNERABILITIES]
        );
    }

    hasVul(artifact: Artifact): boolean {
        return !!(
            artifact &&
            artifact.addition_links &&
            artifact.addition_links[ADDITIONS.VULNERABILITIES]
        );
    }

    submitFinish(e: boolean) {
        this.scanFinishedArtifactLength += 1;
        // all selected scan action has started
        if (this.scanFinishedArtifactLength === this.onScanArtifactsLength) {
            this.onSendingScanCommand = e;
        }
    }

    submitStopFinish(e: boolean) {
        this.scanStoppedArtifactLength += 1;
        // all selected scan action has stopped
        if (this.scanStoppedArtifactLength === this.onStopScanArtifactsLength) {
            this.onSendingScanCommand = e;
        }
    }
    handleScanOverview(scanOverview: any): any {
        if (scanOverview) {
            return Object.values(scanOverview)[0];
        }
        return null;
    }

    goIntoIndexArtifact(artifact: Artifact) {
        let depth: string = '';
        if (this.depth) {
            depth = this.depth + '-' + artifact.digest;
        } else {
            depth = artifact.digest;
        }
        const linkUrl = [
            'harbor',
            'projects',
            this.projectId,
            'repositories',
            this.repoName,
            'artifacts-tab',
            'depth',
            depth,
        ];
        if (this.activatedRoute.snapshot.queryParams[UN_LOGGED_PARAM] === YES) {
            this.router.navigate(linkUrl, {
                queryParams: { [UN_LOGGED_PARAM]: YES },
            });
        } else {
            this.router.navigate(linkUrl);
        }
    }

    selectFilterType() {
        this.lastFilteredTagName = '';
        if (this.filterByType === 'labels') {
            this.openLabelFilterPanel = true;
            // every time  when filer panel opens, resort imageFilterLabels labels
            this.reSortImageFilterLabels();
            this.openLabelFilterPiece = true;
        } else {
            this.openLabelFilterPiece = false;
            this.filterOneLabel = this.initFilter;
            this.showlabel = false;
            this.artifactListPageService.imageFilterLabels.forEach(data => {
                data.iconsShow = false;
            });
        }
        this.currentPage = 1;
        let st: ClrDatagridStateInterface = this.currentState;
        if (!st) {
            st = { page: {} };
        }
        st.page.size = this.pageSize;
        st.page.from = 0;
        st.page.to = this.pageSize - 1;
        this.filters = [];
        this.clrLoad(st);
    }

    selectFilter(showItem: string, filterItem: string) {
        this.lastFilteredTagName = filterItem;
        this.currentPage = 1;

        let st: ClrDatagridStateInterface = this.currentState;
        if (!st) {
            st = { page: {} };
        }
        st.page.size = this.pageSize;
        st.page.from = 0;
        st.page.to = this.pageSize - 1;
        this.filters = [];
        if (filterItem) {
            this.filters.push(`${this.filterByType}=${filterItem}`);
        }

        this.clrLoad(st);
    }

    get isFilterReadonly() {
        return this.filterByType === 'labels' ? 'readonly' : null;
    }

    // when finished, remove it from selectedRow
    scanFinished(artifact: Artifact) {
        if (this.selectedRow && this.selectedRow.length) {
            for (let i = 0; i < this.selectedRow.length; i++) {
                if (artifact.digest === this.selectedRow[i].digest) {
                    this.selectedRow.splice(i, 1);
                    break;
                }
            }
        }
    }

    getIconsFromBackEnd() {
        if (this.artifactList && this.artifactList.length) {
            this.artifactService.getIconsFromBackEnd(this.artifactList);
        }
    }

    showDefaultIcon(event: any) {
        if (event && event.target) {
            event.target.src = artifactDefault;
        }
    }

    getIcon(icon: string): SafeUrl {
        return this.artifactService.getIcon(icon);
    }

    // get Tags and display less than 9 tags(too many tags will make UI stuck)
    getArtifactTagsAsync(artifacts: ArtifactFront[]) {
        if (artifacts && artifacts.length) {
            artifacts.forEach(item => {
                const listTagParams: NewArtifactService.ListTagsParams = {
                    projectName: this.projectName,
                    repositoryName: dbEncodeURIComponent(this.repoName),
                    reference: item.digest,
                    withSignature: true,
                    withImmutableStatus: true,
                    page: 1,
                    pageSize: 8,
                };
                this.newArtifactService
                    .listTagsResponse(listTagParams)
                    .subscribe(res => {
                        if (res.headers) {
                            let xHeader: string =
                                res.headers.get('x-total-count');
                            if (xHeader) {
                                item.tagNumber = Number.parseInt(xHeader, 10);
                            }
                        }
                        item.tags = res.body;
                    });
            });
        }
    }
    // get accessories
    getAccessoriesAsync(artifacts: ArtifactFront[]) {
        if (artifacts && artifacts.length) {
            artifacts.forEach(item => {
                const listTagParams: NewArtifactService.ListAccessoriesParams =
                    {
                        projectName: this.projectName,
                        repositoryName: dbEncodeURIComponent(this.repoName),
                        reference: item.digest,
                        page: 1,
                        pageSize: ACCESSORY_PAGE_SIZE,
                    };
                this.newArtifactService
                    .listAccessoriesResponse(listTagParams)
                    .subscribe(res => {
                        if (res.headers) {
                            let xHeader: string =
                                res.headers.get('x-total-count');
                            if (xHeader) {
                                item.accessoryNumber = Number.parseInt(
                                    xHeader,
                                    10
                                );
                            }
                        }
                        item.accessories = res.body;
                    });
            });
        }
    }
    checkCosignAsync(artifacts: ArtifactFront[]) {
        if (artifacts && artifacts.length) {
            artifacts.forEach(item => {
                item.coSigned = CHECKING;
                const listTagParams: NewArtifactService.ListAccessoriesParams =
                    {
                        projectName: this.projectName,
                        repositoryName: dbEncodeURIComponent(this.repoName),
                        reference: item.digest,
                        q: encodeURIComponent(`type=${AccessoryType.COSIGN}`),
                        page: 1,
                        pageSize: ACCESSORY_PAGE_SIZE,
                    };
                this.newArtifactService
                    .listAccessories(listTagParams)
                    .subscribe(
                        res => {
                            if (res?.length) {
                                item.coSigned = TRUE;
                            } else {
                                item.coSigned = FALSE;
                            }
                        },
                        err => {
                            item.coSigned = FALSE;
                        }
                    );
            });
        }
    }
    // return true if all selected rows are in "running" state
    canStopScan(): boolean {
        if (this.onSendingStopScanCommand) {
            return false;
        }
        if (this.selectedRow && this.selectedRow.length) {
            let flag: boolean = true;
            this.selectedRow.forEach(item => {
                const st: string = this.scanStatus(item);
                if (!this.isRunningState(st)) {
                    flag = false;
                }
            });
            return flag;
        }
        return false;
    }

    isRunningState(state: string): boolean {
        return (
            state === VULNERABILITY_SCAN_STATUS.RUNNING ||
            state === VULNERABILITY_SCAN_STATUS.PENDING ||
            state === VULNERABILITY_SCAN_STATUS.SCHEDULED
        );
    }

    stopNow() {
        if (this.selectedRow && this.selectedRow.length) {
            this.scanStoppedArtifactLength = 0;
            this.onStopScanArtifactsLength = this.selectedRow.length;
            this.onSendingStopScanCommand = true;
            this.selectedRow.forEach((data: any) => {
                let digest = data.digest;
                this.eventService.publish(
                    HarborEvent.STOP_SCAN_ARTIFACT,
                    this.repoName + '/' + digest
                );
            });
        }
    }
    tagsString(tags: Tag[]): string {
        if (tags?.length) {
            const arr: string[] = [];
            tags.forEach(item => {
                arr.push(item.name);
            });
            return arr.join(', ');
        }
        return null;
    }
    deleteAccessory(a: Accessory) {
        let titleKey: string,
            summaryKey: string,
            content: string,
            buttons: ConfirmationButtons;
        titleKey = 'ACCESSORY.DELETION_TITLE_ACCESSORY';
        summaryKey = 'ACCESSORY.DELETION_SUMMARY_ONE_ACCESSORY';
        buttons = ConfirmationButtons.DELETE_CANCEL;
        content = a.digest.slice(0, 15);
        let message = new ConfirmationMessage(
            titleKey,
            summaryKey,
            content,
            a,
            ConfirmationTargets.ACCESSORY,
            buttons
        );
        this.confirmationDialog.open(message);
    }
    isEllipsisActive(ele: HTMLSpanElement): boolean {
        return ele?.offsetWidth < ele?.scrollWidth;
    }
}
