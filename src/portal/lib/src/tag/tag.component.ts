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
  Input,
  Output,
  EventEmitter,
  ChangeDetectorRef,
  ElementRef, AfterViewInit
} from "@angular/core";
import { Subject, forkJoin } from "rxjs";
import { debounceTime, distinctUntilChanged, finalize } from 'rxjs/operators';
import { TranslateService } from "@ngx-translate/core";
import { State, Comparator } from "../service/interface";

import { TagService, RetagService, VulnerabilitySeverity, RequestQueryParams } from "../service/index";
import { ErrorHandler } from "../error-handler/error-handler";
import { ChannelService } from "../channel/index";
import {
  ConfirmationTargets,
  ConfirmationState,
  ConfirmationButtons
} from "../shared/shared.const";

import { ConfirmationDialogComponent } from "../confirmation-dialog/confirmation-dialog.component";
import { ConfirmationMessage } from "../confirmation-dialog/confirmation-message";
import { ConfirmationAcknowledgement } from "../confirmation-dialog/confirmation-state-message";

import { Label, Tag, TagClickEvent, RetagRequest } from "../service/interface";

import {
  CustomComparator,
  calculatePage,
  doFiltering,
  doSorting,
  VULNERABILITY_SCAN_STATUS,
  DEFAULT_PAGE_SIZE,
  clone,
} from "../utils";

import { CopyInputComponent } from "../push-image/copy-input.component";
import { LabelService } from "../service/label.service";
import { UserPermissionService } from "../service/permission.service";
import { USERSTATICPERMISSION } from "../service/permission-static";
import { operateChanges, OperateInfo, OperationState } from "../operation/operate";
import { OperationService } from "../operation/operation.service";
import { ImageNameInputComponent } from "../image-name-input/image-name-input.component";
import { map, catchError } from "rxjs/operators";
import { errorHandler as errorHandFn } from "../shared/shared.utils";
import { Observable, throwError as observableThrowError } from "rxjs";
export interface LabelState {
  iconsShow: boolean;
  label: Label;
  show: boolean;
}
export const AVAILABLE_TIME = '0001-01-01T00:00:00Z';
@Component({
  selector: 'hbr-tag',
  templateUrl: './tag.component.html',
  styleUrls: ['./tag.component.scss']
})
export class TagComponent implements OnInit, AfterViewInit {

  signedCon: { [key: string]: any | string[] } = {};
  @Input() projectId: number;
  @Input() memberRoleID: number;
  @Input() repoName: string;
  @Input() isEmbedded: boolean;

  @Input() hasSignedIn: boolean;
  @Input() isGuest: boolean;
  @Input() registryUrl: string;
  @Input() withNotary: boolean;
  @Input() withClair: boolean;
  @Input() withAdmiral: boolean;
  @Output() refreshRepo = new EventEmitter<boolean>();
  @Output() tagClickEvent = new EventEmitter<TagClickEvent>();
  @Output() signatureOutput = new EventEmitter<any>();


  tags: Tag[];

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

  createdComparator: Comparator<Tag> = new CustomComparator<Tag>("created", "date");

  loading = false;
  copyFailed = false;
  selectedRow: Tag[] = [];

  imageLabels: LabelState[] = [];
  imageStickLabels: LabelState[] = [];
  imageFilterLabels: LabelState[] = [];

  labelListOpen = false;
  selectedTag: Tag[];
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

  @ViewChild("confirmationDialog", {static: false})
  confirmationDialog: ConfirmationDialogComponent;

  @ViewChild("imageNameInput", {static: false})
  imageNameInput: ImageNameInputComponent;

  @ViewChild("digestTarget", {static: false}) textInput: ElementRef;
  @ViewChild("copyInput", {static: false}) copyInput: CopyInputComponent;

  pageSize: number = DEFAULT_PAGE_SIZE;
  currentPage = 1;
  totalCount = 0;
  currentState: State;

  hasAddLabelImagePermission: boolean;
  hasRetagImagePermission: boolean;
  hasDeleteImagePermission: boolean;
  hasScanImagePermission: boolean;
  constructor(
    private errorHandler: ErrorHandler,
    private tagService: TagService,
    private retagService: RetagService,
    private userPermissionService: UserPermissionService,
    private labelService: LabelService,
    private translateService: TranslateService,
    private ref: ChangeDetectorRef,
    private operationService: OperationService,
    private channel: ChannelService
  ) { }

  ngOnInit() {
    if (!this.projectId) {
      this.errorHandler.error("Project ID cannot be unset.");
      return;
    }
    if (!this.repoName) {
      this.errorHandler.error("Repo name cannot be unset.");
      return;
    }
    this.retrieve();
    this.lastFilteredTagName = '';

    this.labelNameFilter
      .pipe(debounceTime(500))
      .pipe(distinctUntilChanged())
      .subscribe((name: string) => {
        if (this.filterName.length) {
          this.filterOnGoing = true;

          this.imageFilterLabels.forEach(data => {
            if (data.label.name.indexOf(this.filterName) !== -1) {
              data.show = true;
            } else {
              data.show = false;
            }
          });
          setTimeout(() => {
            setInterval(() => this.ref.markForCheck(), 200);
          }, 1000);
        }
      });

    this.stickLabelNameFilter
      .pipe(debounceTime(500))
      .pipe(distinctUntilChanged())
      .subscribe((name: string) => {
        if (this.stickName.length) {
          this.filterOnGoing = true;

          this.imageStickLabels.forEach(data => {
            if (data.label.name.indexOf(this.stickName) !== -1) {
              data.show = true;
            } else {
              data.show = false;
            }
          });
          setTimeout(() => {
            setInterval(() => this.ref.markForCheck(), 200);
          }, 1000);
        }
      });

    this.getImagePermissionRule(this.projectId);
  }

  ngAfterViewInit() {
    if (!this.withAdmiral) {
      this.getAllLabels();
    }
  }

  public get filterLabelPieceWidth() {
    let len = this.lastFilteredTagName.length ? this.lastFilteredTagName.length * 6 + 60 : 115;
    return len > 210 ? 210 : len;
  }

  doSearchTagNames(tagName: string) {
    this.lastFilteredTagName = tagName;
    this.currentPage = 1;

    let st: State = this.currentState;
    if (!st) {
      st = { page: {} };
    }
    st.page.size = this.pageSize;
    st.page.from = 0;
    st.page.to = this.pageSize - 1;
    let selectedLab = this.imageFilterLabels.find(label => label.iconsShow === true);
    if (selectedLab) {
      st.filters = [{ property: 'name', value: this.lastFilteredTagName }, { property: 'labels.id', value: selectedLab.label.id }];
    } else {
      st.filters = [{ property: 'name', value: this.lastFilteredTagName }];
    }

    this.clrLoad(st);
  }

  clrLoad(state: State): void {
    this.selectedRow = [];
    // Keep it for future filtering and sorting
    this.currentState = state;

    let pageNumber: number = calculatePage(state);
    if (pageNumber <= 0) { pageNumber = 1; }

    // Pagination
    let params: RequestQueryParams = new RequestQueryParams();
    params = params.set("page", "" + pageNumber).set("page_size", "" + this.pageSize);

    this.loading = true;

    this.tagService.getTags(
      this.repoName,
      params)
      .subscribe((tags: Tag[]) => {
        this.signedCon = {};
        // Do filtering and sorting
        this.tags = doFiltering<Tag>(tags, state);
        this.tags = doSorting<Tag>(this.tags, state);
        this.tags = this.tags.map(tag => {
          tag.push_time = tag.push_time === AVAILABLE_TIME ? '' : tag.push_time;
          return tag;
        });
        this.loading = false;
      }, error => {
        this.loading = false;
        this.errorHandler.error(error);
      });

    // Force refresh view
    let hnd = setInterval(() => this.ref.markForCheck(), 100);
    setTimeout(() => clearInterval(hnd), 5000);
  }

  refresh() {
    this.doSearchTagNames("");
  }

  getAllLabels(): void {
    forkJoin(this.labelService.getGLabels(), this.labelService.getPLabels(this.projectId)).subscribe(results => {
      results.forEach(labels => {
        labels.forEach(data => {
          this.imageLabels.push({ 'iconsShow': false, 'label': data, 'show': true });
        });
      });
      this.imageFilterLabels = clone(this.imageLabels);
      this.imageStickLabels = clone(this.imageLabels);
    }, error => this.errorHandler.error(error));
  }

  labelSelectedChange(tag?: Tag[]): void {
    if (tag && tag[0].labels) {
      this.imageStickLabels.forEach(data => {
        data.iconsShow = false;
        data.show = true;
      });
      if (tag[0].labels.length) {
        tag[0].labels.forEach((labelInfo: Label) => {
          let findedLabel = this.imageStickLabels.find(data => labelInfo.id === data['label'].id);
          this.imageStickLabels.splice(this.imageStickLabels.indexOf(findedLabel), 1);
          this.imageStickLabels.unshift(findedLabel);

          findedLabel.iconsShow = true;
        });
      }
    }
  }

  addLabels(tag: Tag[]): void {
    this.labelListOpen = true;
    this.selectedTag = tag;
    this.stickName = '';
    this.labelSelectedChange(tag);
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
      this.inprogress = true;
      let labelId = labelInfo.label.id;
      this.selectedRow = this.selectedTag;
      this.tagService.addLabelToImages(this.repoName, this.selectedRow[0].name, labelId).subscribe(res => {
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
        this.inprogress = false;
      }, err => {
        this.inprogress = false;
        this.errorHandler.error(err);
      });
    }
  }

  unSelectLabel(labelInfo: LabelState): void {
    if (!this.inprogress) {
      this.inprogress = true;
      let labelId = labelInfo.label.id;
      this.selectedRow = this.selectedTag;
      this.tagService.deleteLabelToImages(this.repoName, this.selectedRow[0].name, labelId).subscribe(res => {
        this.refresh();

        // insert the unselected label to groups with the same icons
        this.sortOperation(this.imageStickLabels, labelInfo);
        labelInfo.iconsShow = false;
        this.inprogress = false;
      }, err => {
        this.inprogress = false;
        this.errorHandler.error(err);
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
    // insert the unselected label to groups with the same icons
    let preLabelInfo = this.imageFilterLabels.find(data => data.label.id === this.filterOneLabel.id);
    if (preLabelInfo) {
      this.sortOperation(this.imageFilterLabels, preLabelInfo);
    }

    this.imageFilterLabels.filter(data => {
      if (data.label.id !== labelId) {
        data.iconsShow = false;
      } else {
        data.iconsShow = true;
      }
    });
    this.imageFilterLabels.splice(this.imageFilterLabels.indexOf(labelInfo), 1);
    this.imageFilterLabels.unshift(labelInfo);
    this.filterOneLabel = labelInfo.label;

    // reload data
    this.currentPage = 1;
    let st: State = this.currentState;
    if (!st) {
      st = { page: {} };
    }
    st.page.size = this.pageSize;
    st.page.from = 0;
    st.page.to = this.pageSize - 1;
    if (this.lastFilteredTagName) {
      st.filters = [{ property: 'name', value: this.lastFilteredTagName }, { property: 'labels.id', value: labelId }];
    } else {
      st.filters = [{ property: 'labels.id', value: labelId }];
    }

    this.clrLoad(st);
  }

  unFilterLabel(labelInfo: LabelState): void {
    // insert the unselected label to groups with the same icons
    this.sortOperation(this.imageFilterLabels, labelInfo);

    this.filterOneLabel = this.initFilter;
    labelInfo.iconsShow = false;

    // reload data
    this.currentPage = 1;
    let st: State = this.currentState;
    if (!st) {
      st = { page: {} };
    }
    st.page.size = this.pageSize;
    st.page.from = 0;
    st.page.to = this.pageSize - 1;
    if (this.lastFilteredTagName) {
      st.filters = [{ property: 'name', value: this.lastFilteredTagName }];
    } else {
      st.filters = [];
    }
    this.clrLoad(st);
  }

  closeFilter(): void {
    this.openLabelFilterPanel = false;
  }

  openFlagEvent(isOpen: boolean): void {
    if (isOpen) {
      this.openLabelFilterPanel = true;
      this.openLabelFilterPiece = true;
      this.filterName = '';
      // redisplay all labels
      this.imageFilterLabels.forEach(data => {
        if (data.label.name.indexOf(this.filterName) !== -1) {
          data.show = true;
        } else {
          data.show = false;
        }
      });
    } else {
      this.openLabelFilterPanel = false;
      this.openLabelFilterPiece = false;
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

  retrieve() {
    this.tags = [];
    let signatures: string[] = [];
    this.loading = true;

    this.tagService
      .getTags(this.repoName)
      .subscribe(items => {
        // To keep easy use for vulnerability bar
        items.forEach((t: Tag) => {
          if (!t.scan_overview) {
            t.scan_overview = {
              scan_status: VULNERABILITY_SCAN_STATUS.stopped,
              severity: VulnerabilitySeverity.UNKNOWN,
              update_time: new Date(),
              components: {
                total: 0,
                summary: []
              }
            };
          }
          if (t.signature !== null) {
            signatures.push(t.name);
          }
        });
        this.tags = items.map(tag => {
          tag.push_time = tag.push_time === AVAILABLE_TIME ? '' : tag.push_time;
          return tag;
        });
        let signedName: { [key: string]: string[] } = {};
        signedName[this.repoName] = signatures;
        this.signatureOutput.emit(signedName);
        this.loading = false;
        if (this.tags && this.tags.length === 0) {
          this.refreshRepo.emit(true);
        }
      }, error => {
        this.errorHandler.error(error);
        this.loading = false;
      });
    let hnd = setInterval(() => this.ref.markForCheck(), 100);
    setTimeout(() => clearInterval(hnd), 5000);
  }

  sizeTransform(tagSize: string): string {
    let size: number = Number.parseInt(tagSize);
    if (Math.pow(1024, 1) <= size && size < Math.pow(1024, 2)) {
      return (size / Math.pow(1024, 1)).toFixed(2) + "KB";
    } else if (Math.pow(1024, 2) <= size && size < Math.pow(1024, 3)) {
      return (size / Math.pow(1024, 2)).toFixed(2) + "MB";
    } else if (Math.pow(1024, 3) <= size && size < Math.pow(1024, 4)) {
      return (size / Math.pow(1024, 3)).toFixed(2) + "GB";
    } else {
      return size + "B";
    }
  }

  retag(tags: Tag[]) {
    if (tags && tags.length) {
      this.retagDialogOpened = true;
      this.retagSrcImage = this.repoName + ":" + tags[0].digest;
    } else {
      this.errorHandler.error("One tag should be selected before retag.");
    }
  }

  onRetag() {
    this.retagService.retag({
      targetProject: this.imageNameInput.projectName.value,
      targetRepo: this.imageNameInput.repoName.value,
      targetTag: this.imageNameInput.tagName.value,
      srcImage: this.retagSrcImage,
      override: true
    })
      .pipe(finalize(() => {
        this.retagDialogOpened = false;
        this.imageNameInput.form.reset();
      }))
      .subscribe(response => {
        this.translateService.get('RETAG.MSG_SUCCESS').subscribe((res: string) => {
          this.errorHandler.info(res);
        });
      }, error => {
        this.errorHandler.error(error);
      });
  }

  deleteTags(tags: Tag[]) {
    if (tags && tags.length) {
      let tagNames: string[] = [];
      tags.forEach(tag => {
        tagNames.push(tag.name);
      });

      let titleKey: string, summaryKey: string, content: string, buttons: ConfirmationButtons;
      titleKey = "REPOSITORY.DELETION_TITLE_TAG";
      summaryKey = "REPOSITORY.DELETION_SUMMARY_TAG";
      buttons = ConfirmationButtons.DELETE_CANCEL;
      content = tagNames.join(" , ");
      let message = new ConfirmationMessage(
        titleKey,
        summaryKey,
        content,
        tags,
        ConfirmationTargets.TAG,
        buttons);
      this.confirmationDialog.open(message);
    }
  }

  confirmDeletion(message: ConfirmationAcknowledgement) {
    if (message &&
      message.source === ConfirmationTargets.TAG
      && message.state === ConfirmationState.CONFIRMED) {
      let tags: Tag[] = message.data;
      if (tags && tags.length) {
        let observableLists: any[] = [];
        tags.forEach(tag => {
          observableLists.push(this.delOperate(tag));
        });

        forkJoin(...observableLists).subscribe((item) => {
          this.selectedRow = [];
          this.retrieve();
        });
      }
    }
  }

  delOperate(tag: Tag): Observable<any> | null {
    // init operation info
    let operMessage = new OperateInfo();
    operMessage.name = 'OPERATION.DELETE_TAG';
    operMessage.data.id = tag.id;
    operMessage.state = OperationState.progressing;
    operMessage.data.name = tag.name;
    this.operationService.publishInfo(operMessage);

    if (tag.signature) {
      forkJoin(this.translateService.get("BATCH.DELETED_FAILURE"),
        this.translateService.get("REPOSITORY.DELETION_SUMMARY_TAG_DENIED")).subscribe(res => {
          let wrongInfo: string = res[1] + "notary -s https://" + this.registryUrl +
            ":4443 -d ~/.docker/trust remove -p " +
            this.registryUrl + "/" + this.repoName +
            " " + name;
          operateChanges(operMessage, OperationState.failure, wrongInfo);
        });
    } else {
      return this.tagService
        .deleteTag(this.repoName, tag.name)
        .pipe(map(
          response => {
            this.translateService.get("BATCH.DELETED_SUCCESS")
              .subscribe(res => {
                operateChanges(operMessage, OperationState.success);
              });
          }), catchError(error => {
            const message = errorHandFn(error);
            this.translateService.get(message).subscribe(res =>
              operateChanges(operMessage, OperationState.failure, res)
            );
            return observableThrowError(message);
          }));
    }
  }

  showDigestId(tag: Tag[]) {
    if (tag && (tag.length === 1)) {
      this.manifestInfoTitle = "REPOSITORY.COPY_DIGEST_ID";
      this.digestId = tag[0].digest;
      this.showTagManifestOpened = true;
      this.copyFailed = false;
    }
  }

  onTagClick(tag: Tag): void {
    if (tag) {
      let evt: TagClickEvent = {
        project_id: this.projectId,
        repository_name: this.repoName,
        tag_name: tag.name
      };
      this.tagClickEvent.emit(evt);
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
  scanStatus(t: Tag): string {
    if (t && t.scan_overview && t.scan_overview.scan_status) {
      return t.scan_overview.scan_status;
    }

    return VULNERABILITY_SCAN_STATUS.unknown;
  }

  existObservablePackage(t: Tag): boolean {
    return t.scan_overview &&
      t.scan_overview.components &&
      t.scan_overview.components.total &&
      t.scan_overview.components.total > 0 ? true : false;
  }

  // Whether show the 'scan now' menu
  canScanNow(t: Tag[]): boolean {
    if (!this.withClair) { return false; }
    if (!this.hasScanImagePermission) { return false; }
    let st: string = this.scanStatus(t[0]);

    return st !== VULNERABILITY_SCAN_STATUS.pending &&
      st !== VULNERABILITY_SCAN_STATUS.running;
  }
  getImagePermissionRule(projectId: number): void {
    let hasAddLabelImagePermission = this.userPermissionService.getPermission(projectId, USERSTATICPERMISSION.REPOSITORY_TAG_LABEL.KEY,
      USERSTATICPERMISSION.REPOSITORY_TAG_LABEL.VALUE.CREATE);
    let hasRetagImagePermission = this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.REPOSITORY.KEY, USERSTATICPERMISSION.REPOSITORY.VALUE.PULL);
    let hasDeleteImagePermission = this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.REPOSITORY_TAG.KEY, USERSTATICPERMISSION.REPOSITORY_TAG.VALUE.DELETE);
    let hasScanImagePermission = this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.KEY, USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.VALUE.CREATE);
    forkJoin(hasAddLabelImagePermission, hasRetagImagePermission, hasDeleteImagePermission, hasScanImagePermission)
      .subscribe(permissions => {
        this.hasAddLabelImagePermission = permissions[0] as boolean;
        this.hasRetagImagePermission = permissions[1] as boolean;
        this.hasDeleteImagePermission = permissions[2] as boolean;
        this.hasScanImagePermission = permissions[3] as boolean;
      }, error => this.errorHandler.error(error));
  }
  // Trigger scan
  scanNow(t: Tag[]): void {
    if (t && t.length) {
      t.forEach((data: any) => {
        let tagId = data.name;
        this.channel.publishScanEvent(this.repoName + "/" + tagId);
      });
    }
  }

  // pull command
  onCpError($event: any): void {
    this.copyInput.setPullCommendShow();
  }
}
