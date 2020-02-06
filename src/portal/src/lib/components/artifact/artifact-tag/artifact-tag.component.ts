
import { Component, OnInit, Input, OnChanges, SimpleChanges, ViewChild, Output, EventEmitter } from '@angular/core';
import { Artifact } from '../artifact';
import { Tag, TagService } from '../../../services';
import { ConfirmationButtons, ConfirmationTargets, ConfirmationState } from '../../../entities/shared.const';
import { ConfirmationMessage, ConfirmationDialogComponent, ConfirmationAcknowledgement } from '../../confirmation-dialog';
import { Observable, of, forkJoin } from 'rxjs';
import { OperateInfo, OperationState, operateChanges } from '../../operation/operate';
import { OperationService } from '../../operation/operation.service';
import { map, catchError } from 'rxjs/operators';
import { TranslateService } from '@ngx-translate/core';
// import { ConfirmationTargets } from 'src/app/shared/shared.const';
import { errorHandler as errorHandFn } from "../../../utils/shared/shared.utils";
@Component({
  selector: 'artifact-tag',
  templateUrl: './artifact-tag.component.html',
  styleUrls: ['./artifact-tag.component.scss']
})

export class ArtifactTagComponent implements OnInit, OnChanges {
  @Input() artifactDetails: Artifact;
  @Input() projectName: string;
  @Input() repositoryName: string;
  @Output() refreshArtifact = new EventEmitter();
  newTagName = {
    name: ''
  };
  selectedRow: Tag[] = [];
  isTagNameExist = false;
  newTagformShow = false;
  loading = false;
  openTag = false;
  @ViewChild("confirmationDialog", { static: false })
  confirmationDialog: ConfirmationDialogComponent;
  constructor(
    private operationService: OperationService,
    private tagService: TagService,
    private translateService: TranslateService,

  ) { }

  ngOnInit() {
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes && changes["artifactDetails"]) {
      // this.originalConfig = clone(this.currentConfig);
    }
  }
  addTag() {
    this.newTagformShow = true;

  }
  cancelAddTag() {
    this.newTagformShow = false;
    this.newTagName = { name: '' };
  }
  saveAddTag() {
    // api
    this.newTagformShow = false;
    this.newTagName = { name: '' };
  }
  removeTag() {
    if (this.selectedRow && this.selectedRow.length) {
      let tagNames: string[] = [];
      this.selectedRow.forEach(artifact => {
        tagNames.push(artifact.name);
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
      this.selectedRow,
      ConfirmationTargets.TAG,
      buttons);
    this.confirmationDialog.open(message);
    }
  }
  confirmDeletion(message: ConfirmationAcknowledgement) {
    if (message &&
      message.source === ConfirmationTargets.TAG
      && message.state === ConfirmationState.CONFIRMED) {
      let tagList: Tag[] = message.data;
      if (tagList && tagList.length) {
        let observableLists: any[] = [];
        tagList.forEach(tag => {
          observableLists.push(this.delOperate(tag));
        });

        forkJoin(...observableLists).subscribe((items) => {
          // if delete one success  refresh list
          if (items.some(item => !item)) {
            this.selectedRow = [];
            this.refreshArtifact.emit();
          }
        });
      }
    }

    if (message &&
      message.source === ConfirmationTargets.TAG
      && message.state === ConfirmationState.CONFIRMED) {
      this.delOperate(message.data).subscribe(res => {
        this.refreshArtifact.emit();
      })

    }
  }

  delOperate(tag): Observable<any> | null {
    // init operation info
    let operMessage = new OperateInfo();
    operMessage.name = 'OPERATION.DELETE_TAG';
    operMessage.state = OperationState.progressing;
    operMessage.data.name = tag.name;
    this.operationService.publishInfo(operMessage);
    return this.tagService
      .deleteTag(this.projectName, this.repositoryName, this.artifactDetails.digest, tag.name)
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
          return of(error);
        }));
    // }
  }

  existValid() {

  }
  toggleTagListOpenOrClose() {
    this.openTag = !this.openTag;
    this.newTagformShow = false;
  }
  // refreshTags(state: ClrDatagridStateInterface) {

  // }
}
