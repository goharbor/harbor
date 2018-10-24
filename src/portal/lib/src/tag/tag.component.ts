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
import {Subject, forkJoin} from "rxjs";
import { debounceTime , distinctUntilChanged} from 'rxjs/operators';
import { TranslateService } from "@ngx-translate/core";
import { State, Comparator } from "@clr/angular";

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
  toPromise,
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
import { operateChanges, OperateInfo, OperationState } from "../operation/operate";
import { OperationService } from "../operation/operation.service";
import { ImageNameInputComponent } from "../image-name-input/image-name-input.component";

export interface LabelState {
  iconsShow: boolean;
  label: Label;
  show: boolean;
}

@Component({
  selector: 'hbr-tag',
  templateUrl: './tag.component.html',
  styleUrls: ['./tag.component.scss']
})
export class TagComponent implements OnInit, AfterViewInit {

  signedCon: {[key: string]: any | string[]} = {};
  @Input() projectId: number;
  @Input() repoName: string;
  @Input() isEmbedded: boolean;

  @Input() hasSignedIn: boolean;
  @Input() hasProjectAdminRole: boolean;
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
  retagDialogClosable = true;
  lastFilteredTagName: string;
  inprogress: boolean;
  openLabelFilterPanel: boolean;
  openLabelFilterPiece: boolean;
  retagSrcImage: string;

  createdComparator: Comparator<Tag> = new CustomComparator<Tag>("created", "date");

  loading = false;
  copyFailed = false;
  selectedRow: Tag[] = [];

  imageLabels: LabelState[] = [];
  imageStickLabels: LabelState[] = [];
  imageFilterLabels: LabelState[] = [];

  labelListOpen = false;
  selectedTag: Tag[];
  labelNameFilter: Subject<string> = new Subject<string> ();
  stickLabelNameFilter: Subject<string> = new Subject<string> ();
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
  currentState: State;

  constructor(
    private errorHandler: ErrorHandler,
    private tagService: TagService,
    private retagService: RetagService,
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
      st.filters = [{property: 'name', value: this.lastFilteredTagName}, {property: 'labels.id', value: selectedLab.label.id}];
    } else {
      st.filters = [{property: 'name', value: this.lastFilteredTagName}];
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
    params.set("page", "" + pageNumber);
    params.set("page_size", "" + this.pageSize);

    this.loading = true;

    toPromise<Tag[]>(this.tagService.getTags(
      this.repoName,
      params))
      .then((tags: Tag[]) => {
        this.signedCon = {};
        // Do filtering and sorting
        this.tags = doFiltering<Tag>(tags, state);
        this.tags = doSorting<Tag>(this.tags, state);

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

  refresh() {
    this.doSearchTagNames("");
  }

  getAllLabels(): void {
    toPromise<Label[]>(this.labelService.getGLabels()).then((res: Label[]) => {
      if (res.length) {
        res.forEach(data => {
          this.imageLabels.push({'iconsShow': false, 'label': data, 'show': true});
        });
      }

      toPromise<Label[]>(this.labelService.getPLabels(this.projectId)).then((res1: Label[]) => {
        if (res1.length) {
          res1.forEach(data => {
            this.imageLabels.push({'iconsShow': false, 'label': data, 'show': true});
          });
        }
        this.imageFilterLabels = clone(this.imageLabels);
        this.imageStickLabels = clone(this.imageLabels);
      }).catch(error => {
        this.errorHandler.error(error);
      });
    }).catch(error => {
      this.errorHandler.error(error);
    });
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
      toPromise<any>(this.tagService.addLabelToImages(this.repoName, this.selectedRow[0].name, labelId)).then(res => {
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
      }).catch(err => {
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
        toPromise<any>(this.tagService.deleteLabelToImages(this.repoName, this.selectedRow[0].name, labelId)).then(res => {
          this.refresh();

          // insert the unselected label to groups with the same icons
          this.sortOperation(this.imageStickLabels, labelInfo);
        labelInfo.iconsShow = false;
        this.inprogress = false;
      }).catch(err => {
        this.inprogress = false;
        this.errorHandler.error(err);
      });
    }
  }

  rightFilterLabel(labelInfo: LabelState): void {
    if (labelInfo) {
      if (!labelInfo.iconsShow) {
        this.filterLabel(labelInfo);
      } else {
        this.unFilterLabel(labelInfo);
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
        st.filters = [{property: 'name', value: this.lastFilteredTagName}, {property: 'labels.id', value: labelId}];
      } else {
        st.filters = [{property: 'labels.id', value: labelId}];
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
      st.filters = [{property: 'name', value: this.lastFilteredTagName}];
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
    } else  {
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
    let signatures: string[] = [] ;
    this.loading = true;

    toPromise<Tag[]>(this.tagService
      .getTags(this.repoName))
      .then(items => {
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
      this.tags = items;
        let signedName: {[key: string]: string[]} = {};
        signedName[this.repoName] = signatures;
        this.signatureOutput.emit(signedName);
        this.loading = false;
        if (this.tags && this.tags.length === 0) {
          this.refreshRepo.emit(true);
        }
      })
      .catch(error => {
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
      return  (size / Math.pow(1024, 2)).toFixed(2) + "MB";
    } else if (Math.pow(1024, 3) <= size && size < Math.pow(1024, 4)) {
      return  (size / Math.pow(1024, 3)).toFixed(2) + "GB";
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
    this.retagDialogOpened = false;
    this.retagService.retag({
        targetProject: this.imageNameInput.projectName.value,
        targetRepo: this.imageNameInput.repoName.value,
        targetTag: this.imageNameInput.tagName.value,
        srcImage: this.retagSrcImage,
        override: true
     }).subscribe(response => {
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
        let promiseLists: any[] = [];
        tags.forEach(tag => {
          promiseLists.push(this.delOperate(tag));
        });

        Promise.all(promiseLists).then((item) => {
          this.selectedRow = [];
          this.retrieve();
        });
      }
    }
  }

  delOperate(tag: Tag) {
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
      return toPromise<number>(this.tagService
          .deleteTag(this.repoName, tag.name))
          .then(
              response => {
                this.translateService.get("BATCH.DELETED_SUCCESS")
                    .subscribe(res =>  {
                      operateChanges(operMessage, OperationState.success);
                    });
              }).catch(error => {
            if (error.status === 503) {
              forkJoin(this.translateService.get('BATCH.DELETED_FAILURE'),
                  this.translateService.get('REPOSITORY.TAGS_NO_DELETE')).subscribe(res => {
                operateChanges(operMessage, OperationState.failure, res[1]);
              });
              return;
            }
            this.translateService.get("BATCH.DELETED_FAILURE").subscribe(res => {
              operateChanges(operMessage, OperationState.failure, res);
            });
          });
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
    if (!this.hasProjectAdminRole) { return false; }
      let st: string = this.scanStatus(t[0]);

    return st !== VULNERABILITY_SCAN_STATUS.pending &&
      st !== VULNERABILITY_SCAN_STATUS.running;
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
