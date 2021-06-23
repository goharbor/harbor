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
import { Component, ElementRef, Input, OnDestroy, OnInit, ViewChild, } from "@angular/core";
import { forkJoin, Observable, of, Subject, Subscription } from "rxjs";
import { catchError, debounceTime, distinctUntilChanged, finalize, map } from 'rxjs/operators';
import { TranslateService } from "@ngx-translate/core";
import { ClrDatagridComparatorInterface, ClrDatagridStateInterface, ClrLoadingState } from "@clr/angular";

import { ActivatedRoute, Router } from "@angular/router";
import { Comparator, ScanningResultService, UserPermissionService, USERSTATICPERMISSION, } from "../../../../../../../shared/services";
import {
  calculatePage,
  clone,
  CustomComparator,
  dbEncodeURIComponent,
  DEFAULT_PAGE_SIZE,
  DEFAULT_SUPPORTED_MIME_TYPES,
  doSorting,
  formatSize,
  getSortingString,
  VULNERABILITY_SCAN_STATUS
} from "../../../../../../../shared/units/utils";
import { ImageNameInputComponent } from "../../../../../../../shared/components/image-name-input/image-name-input.component";
import { CopyInputComponent } from "../../../../../../../shared/components/push-image/copy-input.component";
import { ErrorHandler } from "../../../../../../../shared/units/error-handler";
import { ArtifactService } from "../../../artifact.service";
import { OperationService } from "../../../../../../../shared/components/operation/operation.service";
import { ChannelService } from "../../../../../../../shared/services/channel.service";
import { ConfirmationButtons, ConfirmationState, ConfirmationTargets } from "../../../../../../../shared/entities/shared.const";
import { operateChanges, OperateInfo, OperationState } from "../../../../../../../shared/components/operation/operate";
import { artifactDefault, ArtifactFront as Artifact, ArtifactFront, artifactPullCommands, mutipleFilter } from '../../../artifact';
import { Project } from "../../../../../project";
import { ArtifactService as NewArtifactService } from "../../../../../../../../../ng-swagger-gen/services/artifact.service";
import { ADDITIONS } from "../../../artifact-additions/models";
import { Platform } from "../../../../../../../../../ng-swagger-gen/models/platform";
import { SafeUrl } from '@angular/platform-browser';
import { errorHandler } from "../../../../../../../shared/units/shared.utils";
import { ConfirmationDialogComponent } from "../../../../../../../shared/components/confirmation-dialog";
import { ConfirmationMessage } from "../../../../../../global-confirmation-dialog/confirmation-message";
import { ConfirmationAcknowledgement } from "../../../../../../global-confirmation-dialog/confirmation-state-message";
import { UN_LOGGED_PARAM } from "../../../../../../../account/sign-in/sign-in.service";
import { Label } from "../../../../../../../../../ng-swagger-gen/models/label";
import { LabelService } from "../../../../../../../../../ng-swagger-gen/services/label.service";

export interface LabelState {
  iconsShow: boolean;
  label: Label;
  show: boolean;
}
export const AVAILABLE_TIME = '0001-01-01T00:00:00.000Z';
const YES: string = 'yes';
const PAGE_SIZE: number = 100;
@Component({
  selector: 'artifact-list-tab',
  templateUrl: './artifact-list-tab.component.html',
  styleUrls: ['./artifact-list-tab.component.scss']
})
export class ArtifactListTabComponent implements OnInit, OnDestroy {

  signedCon: { [key: string]: any | string[] } = {};
  @Input() projectId: number;
  projectName: string;
  @Input() memberRoleID: number;
  @Input() repoName: string;
  @Input() isEmbedded: boolean;
  @Input() hasSignedIn: boolean;
  @Input() isGuest: boolean;
  @Input() registryUrl: string;
  @Input() withNotary: boolean;
  @Input() withAdmiral: boolean;
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

  pullComparator: Comparator<Artifact> = new CustomComparator<Artifact>("pull_time", "date");
  pushComparator: Comparator<Artifact> = new CustomComparator<Artifact>("push_time", "date");

  loading = true;
  copyFailed = false;
  selectedRow: Artifact[] = [];

  imageLabels: LabelState[] = [];
  imageStickLabels: LabelState[] = [];
  imageFilterLabels: LabelState[] = [];

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

  @ViewChild("confirmationDialog")
  confirmationDialog: ConfirmationDialogComponent;

  @ViewChild("imageNameInput")
  imageNameInput: ImageNameInputComponent;

  @ViewChild("digestTarget") textInput: ElementRef;
  @ViewChild("copyInput") copyInput: CopyInputComponent;

  pageSize: number = DEFAULT_PAGE_SIZE;
  currentPage = 1;
  totalCount = 0;
  currentState: ClrDatagridStateInterface;

  hasAddLabelImagePermission: boolean;
  hasRetagImagePermission: boolean;
  hasDeleteImagePermission: boolean;
  hasScanImagePermission: boolean;
  hasEnabledScanner: boolean;
  scanBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
  onSendingScanCommand: boolean;

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
  constructor(
    private errorHandlerService: ErrorHandler,
    private userPermissionService: UserPermissionService,
    private labelService: LabelService,
    private artifactService: ArtifactService,
    private newArtifactService: NewArtifactService,
    private translateService: TranslateService,
    private operationService: OperationService,
    private channel: ChannelService,
    private activatedRoute: ActivatedRoute,
    private scanningService: ScanningResultService,
    private router:  Router,
  ) {
  }
  ngOnInit() {
    this.activatedRoute.params.subscribe(params => {
      this.depth = this.activatedRoute.snapshot.params['depth'];
      if (this.depth) {
        const arr: string[] = this.depth.split('-');
        this.artifactDigest = this.depth.split('-')[arr.length - 1];
      }
      if (this.hasInit) {
        this.currentPage = 1;
        this.totalCount = 0;
        const st: ClrDatagridStateInterface = {page: {from: 0, to: this.pageSize - 1, size: this.pageSize}};
        this.clrLoad(st);
      }
      this.init();
    });
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
  }
  init() {
    this.hasInit = true;
    this.depth = this.activatedRoute.snapshot.params['depth'];
    if (this.depth) {
      const arr: string[] = this.depth.split('-');
      this.artifactDigest = this.depth.split('-')[arr.length - 1];
    }
    if (!this.projectId) {
      this.errorHandlerService.error("Project ID cannot be unset.");
      return;
    }
    const resolverData = this.activatedRoute.snapshot.data;
    if (resolverData) {
      const pro: Project = <Project>resolverData['projectResolver'];
      this.projectName = pro.name;
    }

    this.getProjectScanner();
    if (!this.repoName) {
      this.errorHandlerService.error("Repo name cannot be unset.");
      return;
    }
    if (!this.triggerSub) {
      this.triggerSub = this.artifactService.TriggerArtifactChan$.subscribe(res => {
        let st: ClrDatagridStateInterface = { page: {from: 0, to: this.pageSize - 1, size: this.pageSize} };
        this.clrLoad(st);
      });
    }
    this.lastFilteredTagName = '';
    if (!this.labelNameFilterSub) {
      this.labelNameFilterSub = this.labelNameFilter
        .pipe(debounceTime(500))
        .pipe(distinctUntilChanged())
        .subscribe((name: string) => {
          if (this.filterName.length) {
            this.filterOnGoing = true;
            this.imageFilterLabels.forEach(data => {
              data.show = data.label.name.indexOf(this.filterName) !== -1;
            });
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
            this.imageStickLabels.forEach(data => {
              data.show = data.label.name.indexOf(this.stickName) !== -1;
            });
          }
        });
    }
    this.getImagePermissionRule(this.projectId);
  }

  public get filterLabelPieceWidth() {
    let len = this.lastFilteredTagName.length ? this.lastFilteredTagName.length * 6 + 60 : 115;
    return len > 210 ? 210 : len;
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
      this.filters.push(`${this.filterByType}=~${this.lastFilteredTagName}`);
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
      this.selectedRow = [];
      // Keep it for future filtering and sorting

      let pageNumber: number = calculatePage(state);
      if (pageNumber <= 0) { pageNumber = 1; }
      let sortBy: any = '';
      if (state.sort) {
        sortBy = state.sort.by as string | ClrDatagridComparatorInterface<any>;
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
        let q = "";
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
          XAcceptVulnerabilities: DEFAULT_SUPPORTED_MIME_TYPES
        };
        this.newArtifactService.getArtifact(artifactParam).subscribe(
          res => {
            let observableLists: Observable<Artifact>[] = [];
            let platFormAttr: { platform: Platform }[] = [];
            this.totalCount = res.references.length;
            res.references.forEach((child, index) => {
              if (index >= (pageNumber - 1) * this.pageSize && index < pageNumber * this.pageSize) {
                let childParams: NewArtifactService.GetArtifactParams = {
                  repositoryName: dbEncodeURIComponent(this.repoName),
                  projectName: this.projectName,
                  reference: child.child_digest,
                  withImmutableStatus: true,
                  withLabel: true,
                  withScanOverview: true,
                  withTag: false,
                  XAcceptVulnerabilities: DEFAULT_SUPPORTED_MIME_TYPES
                };
                platFormAttr.push({platform: child.platform});
                observableLists.push(this.newArtifactService.getArtifact(childParams));
              }
            });
            forkJoin(observableLists).pipe(finalize(() => {
              this.loading = false;
            })).subscribe(artifacts => {
              this.artifactList = artifacts;
              this.artifactList = doSorting<ArtifactFront>(this.artifactList, state);
              this.artifactList.forEach((artifact, index) => {
                artifact.platform = clone(platFormAttr[index].platform);
              });
              this.getPullCommand(this.artifactList);
              this.getArtifactTagsAsync(this.artifactList);
              this.getIconsFromBackEnd();
            }, error => {
              this.errorHandlerService.error(error);
            });
          }, error => {
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
          XAcceptVulnerabilities: DEFAULT_SUPPORTED_MIME_TYPES
        };
        Object.assign(listArtifactParams, params);
        this.newArtifactService.listArtifactsResponse(listArtifactParams)
          .pipe(finalize(() => this.loading = false))
          .subscribe(res => {
            if (res.headers) {
              let xHeader: string = res.headers.get("X-Total-Count");
              if (xHeader) {
                this.totalCount = parseInt(xHeader, 0);
              }
            }
            this.artifactList = res.body;
            this.getPullCommand(this.artifactList);
            this.getArtifactTagsAsync(this.artifactList);
            this.getIconsFromBackEnd();
          }, error => {
            // error
            this.errorHandlerService.error(error);
          });
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

  getPullCommand(artifactList: Artifact[]) {
    artifactList.forEach(artifact => {
      artifact.pullCommand = '';
      artifactPullCommands.forEach(artifactPullCommand => {
        if (artifactPullCommand.type === artifact.type) {
          artifact.pullCommand =
          `${artifactPullCommand.pullCommand} ${this.registryUrl ?
            this.registryUrl : location.hostname}/${this.projectName}/${this.repoName}@${artifact.digest}`;
        }
      });
    });
  }
  getAllLabels(): void {
    // get all project labels
    this.labelService.ListLabelsResponse({
      pageSize: PAGE_SIZE,
      page: 1,
      scope: 'p',
      projectId: this.projectId
    }).subscribe(res => {
      if (res.headers) {
        const xHeader: string = res.headers.get("X-Total-Count");
        const totalCount = parseInt(xHeader, 0);
        let arr = res.body || [];
        if (totalCount <= PAGE_SIZE) { // already gotten all project labels
          if (arr && arr.length) {
            arr.forEach(data => {
              this.imageLabels.push({ 'iconsShow': false, 'label': data, 'show': true });
            });
            this.imageFilterLabels = clone(this.imageLabels);
            this.imageStickLabels = clone(this.imageLabels);
          }
        } else { // get all the project labels in specified times
          const times: number = Math.ceil(totalCount / PAGE_SIZE);
          const observableList: Observable<Label[]>[] = [];
          for (let i = 2; i <= times; i++) {
            observableList.push(this.labelService.ListLabels({
              page: i,
              pageSize: PAGE_SIZE,
              scope: 'p',
              projectId: this.projectId
            }));
          }
          this.handleLabelRes(observableList, arr);
        }
      }
    });
    // get all global labels
    this.labelService.ListLabelsResponse({
      pageSize: PAGE_SIZE,
      page: 1,
      scope: 'g',
    }).subscribe(res => {
      if (res.headers) {
        const xHeader: string = res.headers.get("X-Total-Count");
        const totalCount = parseInt(xHeader, 0);
        let arr = res.body || [];
        if (totalCount <= PAGE_SIZE) { // already gotten all global labels
          if (arr && arr.length) {
            arr.forEach(data => {
              this.imageLabels.push({ 'iconsShow': false, 'label': data, 'show': true });
            });
            this.imageFilterLabels = clone(this.imageLabels);
            this.imageStickLabels = clone(this.imageLabels);
          }
        } else { // get all the global labels in specified times
          const times: number = Math.ceil(totalCount / PAGE_SIZE);
          const observableList: Observable<Label[]>[] = [];
          for (let i = 2; i <= times; i++) {
            observableList.push(this.labelService.ListLabels({
              page: i,
              pageSize: PAGE_SIZE,
              scope: 'g',
            }));
          }
          this.handleLabelRes(observableList, arr);
        }
      }
    });
  }
  handleLabelRes(observableList: Observable<Label[]>[], arr: Label[]) {
    forkJoin(observableList).subscribe(response => {
      if (response && response.length) {
        response.forEach(item => {
          arr = arr.concat(item);
        });
        arr.forEach(data => {
          this.imageLabels.push({ 'iconsShow': false, 'label': data, 'show': true });
        });
        this.imageFilterLabels = clone(this.imageLabels);
        this.imageStickLabels = clone(this.imageLabels);
      }
    });
  }

  labelSelectedChange(artifact?: Artifact[]): void {
    this.imageStickLabels.forEach(data => {
      data.iconsShow = false;
      data.show = true;
    });
    if (artifact && artifact[0].labels && artifact[0].labels.length) {
      artifact[0].labels.forEach((labelInfo: Label) => {
        let findedLabel = this.imageStickLabels.find(data => labelInfo.id === data['label'].id);
        if (findedLabel) {
          this.imageStickLabels.splice(this.imageStickLabels.indexOf(findedLabel), 1);
          this.imageStickLabels.unshift(findedLabel);
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
      for (let i = 0; i < this.selectedRow.length; i ++) {
        if (this.selectedRow[i].labels && this.selectedRow[i].labels.length) {
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
    if (!this.inprogress) { // add label to multiple artifact
      const ObservableArr: Array<Observable<null>> = [];
      this.selectedRow.forEach(item => {
        const params: NewArtifactService.AddLabelParams = {
          projectName: this.projectName,
          repositoryName: dbEncodeURIComponent(this.repoName),
          reference: item.digest,
          label: labelInfo.label
        };
        ObservableArr.push(this.newArtifactService.addLabel(params));
      });
      this.inprogress = true;
      forkJoin(ObservableArr)
          .pipe(finalize(() => this.inprogress = false))
          .subscribe(res => {
            this.refresh();
            // set the selected label in front
            this.imageStickLabels.splice(this.imageStickLabels.indexOf(labelInfo), 1);
            this.imageStickLabels.some((data, i) => {
              if (!data.iconsShow) {
                this.imageStickLabels.splice(i, 0, labelInfo);
                return true;
              }
            });
            // when is the last one
            if (this.imageStickLabels.every(data => data.iconsShow === true)) {
              this.imageStickLabels.push(labelInfo);
            }
            labelInfo.iconsShow = true;
          }, err => {
            this.refresh();
            this.errorHandlerService.error(err);
          });
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
        labelId: labelId
      };
      this.newArtifactService.removeLabel(params).subscribe(res => {
        this.refresh();

        // insert the unselected label to groups with the same icons
        this.sortOperation(this.imageStickLabels, labelInfo);
        labelInfo.iconsShow = false;
        this.inprogress = false;
      }, err => {
        this.inprogress = false;
        this.errorHandlerService.error(err);
      });
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
    this.imageFilterLabels.filter(data => {
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
    if (this.imageFilterLabels && this.imageFilterLabels.length) {
      for (let i = 0; i < this.imageFilterLabels.length; i++) {
        if (this.imageFilterLabels[i].iconsShow) {
          const arr: LabelState[] = this.imageFilterLabels.splice(i, 1);
          this.imageFilterLabels.unshift(...arr);
          break;
        }
      }
    }
  }
  getFilterPlaceholder(): string {
    return this.showlabel ? "" : 'ARTIFACT.FILTER_FOR_ARTIFACTS';
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
      this.imageFilterLabels.forEach(data => {
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
      this.imageFilterLabels.every(data => data.show = true);
    }
  }

  handleStickInputFilter() {
    if (this.stickName.length) {
      this.stickLabelNameFilter.next(this.stickName);
    } else {
      this.imageStickLabels.every(data => data.show = true);
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
        if (data.label.scope !== labelInfo.label.scope && i === labelList.length - 1) {
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
      this.retagSrcImage = this.repoName + ":" + this.selectedRow[0].digest;
    }
  }

  onRetag() {
    let params: NewArtifactService.CopyArtifactParams = {
      projectName: this.imageNameInput.projectName.value,
      repositoryName: dbEncodeURIComponent(this.imageNameInput.repoName.value),
      from: `${this.projectName}/${this.repoName}@${this.selectedRow[0].digest}`,
    };
    this.newArtifactService.CopyArtifact(params)
      .pipe(finalize(() => {
        this.imageNameInput.form.reset();
        this.retagDialogOpened = false;
      }))
      .subscribe(response => {
        this.translateService.get('RETAG.MSG_SUCCESS').subscribe((res: string) => {
          this.errorHandlerService.info(res);
        });
      }, error => {
        this.errorHandlerService.error(error);
      });
  }

  deleteArtifact() {
    if (this.selectedRow && this.selectedRow.length && !this.depth) {
      let artifactNames: string[] = [];
      this.selectedRow.forEach(artifact => {
        artifactNames.push(artifact.digest.slice(0, 15));
      });

      let titleKey: string, summaryKey: string, content: string, buttons: ConfirmationButtons;
      titleKey = "REPOSITORY.DELETION_TITLE_ARTIFACT";
      summaryKey = "REPOSITORY.DELETION_SUMMARY_ARTIFACT";
      buttons = ConfirmationButtons.DELETE_CANCEL;
      content = artifactNames.join(" , ");
      let message = new ConfirmationMessage(
        titleKey,
        summaryKey,
        content,
        this.selectedRow,
        ConfirmationTargets.TAG,
        buttons);
      this.confirmationDialog.open(message);
    }
  }
  deleteArtifactobservableLists: Observable<any>[] = [];
  confirmDeletion(message: ConfirmationAcknowledgement) {
    if (message &&
      message.source === ConfirmationTargets.TAG
      && message.state === ConfirmationState.CONFIRMED) {
      let artifactList = message.data;
      if (artifactList && artifactList.length) {
        artifactList.forEach(artifact => {
          this.deleteArtifactobservableLists.push(this.delOperate(artifact));
        });
        this.loading = true;
        forkJoin(...this.deleteArtifactobservableLists).subscribe((deleteResult) => {
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
            let st: ClrDatagridStateInterface = { page: {from: 0, to: this.pageSize - 1, size: this.pageSize} };
            this.clrLoad(st);
          } else if (deleteErrorList.length === deleteResult.length) {
            // all is error
            this.loading = false;
            this.errorHandlerService.error(deleteResult[deleteResult.length - 1]);
          } else {
            // some artifact delete success but it has error delete things
            this.errorHandlerService.error(deleteErrorList[deleteErrorList.length - 1]);
            // if delete one success  refresh list
            let st: ClrDatagridStateInterface = { page: {from: 0, to: this.pageSize - 1, size: this.pageSize} };
            this.clrLoad(st);
          }
        });
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
    // to do signature
    // if (tag.signature) {
    //   forkJoin(this.translateService.get("BATCH.DELETED_FAILURE"),
    //     this.translateService.get("REPOSITORY.DELETION_SUMMARY_TAG_DENIED")).subscribe(res => {
    //       let wrongInfo: string = res[1] + "notary -s https://" + this.registryUrl +
    //         ":4443 -d ~/.docker/trust remove -p " +
    //         this.registryUrl + "/" + this.repoName +
    //         " " + name;
    //       operateChanges(operMessage, OperationState.failure, wrongInfo);
    //     });
    // } else {
      let params: NewArtifactService.DeleteArtifactParams = {
        projectName: this.projectName,
        repositoryName: dbEncodeURIComponent(this.repoName),
        reference: artifact.digest
      };
    return this.newArtifactService
      .deleteArtifact(params)
      .pipe(map(
        response => {
          this.translateService.get("BATCH.DELETED_SUCCESS")
            .subscribe(res => {
              operateChanges(operMessage, OperationState.success);
            });
        }), catchError(error => {
          const message = errorHandler(error);
          this.translateService.get(message).subscribe(res =>
            operateChanges(operMessage, OperationState.failure, res)
          );
          return of(error);
        }));
    // }
  }

  showDigestId() {
    if (this.selectedRow && (this.selectedRow.length === 1) && !this.depth) {
      this.manifestInfoTitle = "REPOSITORY.COPY_DIGEST_ID";
      this.digestId = this.selectedRow[0].digest;
      this.showTagManifestOpened = true;
      this.copyFailed = false;
    }
  }

  goIntoArtifactSummaryPage(artifact: Artifact): void {
    const relativeRouterLink: string[] = ['artifacts', artifact.digest];
    if (this.activatedRoute.snapshot.queryParams[UN_LOGGED_PARAM] === YES) {
      this.router.navigate(relativeRouterLink , { relativeTo: this.activatedRoute, queryParams: {[UN_LOGGED_PARAM]: YES} });
    } else {
      this.router.navigate(relativeRouterLink , { relativeTo: this.activatedRoute });
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
  // Whether show the 'scan now' menu
  canScanNow(): boolean {
    if (!this.hasScanImagePermission) { return false; }
    if (this.onSendingScanCommand) { return false; }
    let st: string = this.scanStatus(this.selectedRow[0]);
    return st !== VULNERABILITY_SCAN_STATUS.RUNNING;
  }
  getImagePermissionRule(projectId: number): void {
    const permissions = [
      { resource: USERSTATICPERMISSION.REPOSITORY_ARTIFACT_LABEL.KEY, action: USERSTATICPERMISSION.REPOSITORY_ARTIFACT_LABEL.VALUE.CREATE },
      { resource: USERSTATICPERMISSION.REPOSITORY.KEY, action: USERSTATICPERMISSION.REPOSITORY.VALUE.PULL },
      { resource: USERSTATICPERMISSION.ARTIFACT.KEY, action: USERSTATICPERMISSION.ARTIFACT.VALUE.DELETE },
      { resource: USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.KEY, action: USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.VALUE.CREATE },
    ];
    this.userPermissionService.hasProjectPermissions(this.projectId, permissions).subscribe((results: Array<boolean>) => {
      this.hasAddLabelImagePermission = results[0];
      this.hasRetagImagePermission = results[1];
      this.hasDeleteImagePermission = results[2];
      this.hasScanImagePermission = results[3];
      // only has label permission
      if (this.hasAddLabelImagePermission) {
        if (!this.withAdmiral) {
          this.getAllLabels();
        }
      }
    }, error => this.errorHandlerService.error(error));
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
      this.channel.publishScanEvent(this.repoName + "/" + digest);
    });
  }
  selectedRowHasVul(): boolean {
    return !!(this.selectedRow
      && this.selectedRow[0]
      && this.selectedRow[0].addition_links
      && this.selectedRow[0].addition_links[ADDITIONS.VULNERABILITIES]);
  }
  hasVul(artifact: Artifact): boolean {
    return !!(artifact && artifact.addition_links && artifact.addition_links[ADDITIONS.VULNERABILITIES]);
  }
  submitFinish(e: boolean) {
    this.scanFinishedArtifactLength += 1;
    // all selected scan action has start
    if (this.scanFinishedArtifactLength === this.onScanArtifactsLength) {
      this.onSendingScanCommand = e;
    }
  }
  // pull command
  onCpError($event: any): void {
    this.copyInput.setPullCommendShow();
  }
  getProjectScanner(): void {
    this.hasEnabledScanner = false;
    this.scanBtnState = ClrLoadingState.LOADING;
    this.scanningService.getProjectScanner(this.projectId)
      .subscribe(response => {
        if (response && "{}" !== JSON.stringify(response) && !response.disabled
          && response.health === "healthy") {
          this.scanBtnState = ClrLoadingState.SUCCESS;
          this.hasEnabledScanner = true;
        } else {
          this.scanBtnState = ClrLoadingState.ERROR;
        }
      }, error => {
        this.scanBtnState = ClrLoadingState.ERROR;
      });
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
    const linkUrl = ['harbor', 'projects', this.projectId, 'repositories', this.repoName, 'depth', depth];
    if (this.activatedRoute.snapshot.queryParams[UN_LOGGED_PARAM] === YES) {
      this.router.navigate(linkUrl, {queryParams: {[UN_LOGGED_PARAM]: YES}});
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
      this.imageFilterLabels.forEach(data => {
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
      for ( let i = 0; i < this.selectedRow.length; i++) {
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
          pageSize: 8
        };
        this.newArtifactService.listTagsResponse(listTagParams).subscribe(
            res => {
              if (res.headers) {
                let xHeader: string = res.headers.get("x-total-count");
                if (xHeader) {
                  item.tagNumber = Number.parseInt(xHeader);
                }
              }
              item.tags = res.body;
            }
        );
      });
    }
  }
}
