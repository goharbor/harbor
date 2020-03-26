import { Component, OnInit, Input, ViewChild, Output, EventEmitter } from '@angular/core';
import { Observable, of, forkJoin } from 'rxjs';
import { map, catchError } from 'rxjs/operators';
import { TranslateService } from '@ngx-translate/core';
import { NgForm } from '@angular/forms';
import { AVAILABLE_TIME } from "../../artifact-list-page/artifact-list/artifact-list-tab/artifact-list-tab.component";
import {
  ConfirmationAcknowledgement,
  ConfirmationDialogComponent,
  ConfirmationMessage
} from "../../../../../lib/components/confirmation-dialog";
import { OperationService } from "../../../../../lib/components/operation/operation.service";
import { ErrorHandler } from "../../../../../lib/utils/error-handler";
import { ConfirmationButtons, ConfirmationState, ConfirmationTargets } from "../../../../../lib/entities/shared.const";
import { operateChanges, OperateInfo, OperationState } from "../../../../../lib/components/operation/operate";
import { errorHandler } from "../../../../../lib/utils/shared/shared.utils";
import { ArtifactFront as Artifact } from "../artifact";
import { ArtifactService } from '../../../../../../ng-swagger-gen/services/artifact.service';
import { Tag } from '../../../../../../ng-swagger-gen/models/tag';

class InitTag {
  name = "";
}
@Component({
  selector: 'artifact-tag',
  templateUrl: './artifact-tag.component.html',
  styleUrls: ['./artifact-tag.component.scss']
})

export class ArtifactTagComponent implements OnInit {
  @Input() artifactDetails: Artifact;
  @Input() projectName: string;
  @Input() repositoryName: string;
  @Output() refreshArtifact = new EventEmitter();
  newTagName = new InitTag();
  newTagForm: NgForm;
  @ViewChild("newTagForm", { static: true }) currentForm: NgForm;
  selectedRow: Tag[] = [];
  isTagNameExist = false;
  newTagformShow = false;
  loading = false;
  openTag = false;
  availableTime = AVAILABLE_TIME;
  @ViewChild("confirmationDialog", { static: false })
  confirmationDialog: ConfirmationDialogComponent;
  constructor(
    private operationService: OperationService,
    private artifactService: ArtifactService,
    private translateService: TranslateService,
    private errorHandlerService: ErrorHandler

  ) { }

  ngOnInit() {
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
      repositoryName: this.repositoryName,
      reference: this.artifactDetails.digest,
      tag:  this.newTagName
    };
    this.artifactService.createTag(createTagParams).subscribe(res => {
      this.newTagformShow = false;
      this.newTagName = new InitTag();
      this.refreshArtifact.emit();
    }, error => {
      this.errorHandlerService.error(error);
    });
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
  }

  delOperate(tag): Observable<any> | null {
    // init operation info
    let operMessage = new OperateInfo();
    operMessage.name = 'OPERATION.DELETE_TAG';
    operMessage.state = OperationState.progressing;
    operMessage.data.name = tag.name;
    this.operationService.publishInfo(operMessage);
     const deleteTagParams: ArtifactService.DeleteTagParams = {
      projectName: this.projectName,
      repositoryName: this.repositoryName,
      reference: this.artifactDetails.digest,
      tagName: tag.name
    };
    return this.artifactService.deleteTag(deleteTagParams)
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
  }

  existValid(name) {
    this.isTagNameExist = false;
    if (this.artifactDetails.tags) {
      this.artifactDetails.tags.forEach(tag => {
        if (tag.name === name) {
          this.isTagNameExist = true;
        }
      });
    }

  }
  toggleTagListOpenOrClose() {
    this.openTag = !this.openTag;
    this.newTagformShow = false;
  }
}
