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
import { Component, OnInit, ViewChild, Input, ChangeDetectionStrategy, ChangeDetectorRef } from '@angular/core';

import { TagService } from '../service/tag.service';
import { ErrorHandler } from '../error-handler/error-handler';
import { ConfirmationTargets, ConfirmationState, ConfirmationButtons } from '../shared/shared.const';

import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ConfirmationMessage } from '../confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../confirmation-dialog/confirmation-state-message';

import { Tag, SessionInfo } from '../service/interface';

import { TAG_TEMPLATE } from './tag.component.html';
import { TAG_STYLE } from './tag.component.css';

import { toPromise, CustomComparator } from '../utils';

import { TranslateService } from '@ngx-translate/core';

import { State, Comparator } from 'clarity-angular';

@Component({
  selector: 'hbr-tag',
  template: TAG_TEMPLATE,
  styles: [ TAG_STYLE ],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class TagComponent implements OnInit {

  @Input() projectId: number;
  @Input() repoName: string;
  @Input() sessionInfo: SessionInfo;  

  hasProjectAdminRole: boolean;

  tags: Tag[];

  registryUrl: string;
  withNotary: boolean;
  hasSignedIn: boolean;

  showTagManifestOpened: boolean;
  manifestInfoTitle: string;
  tagID: string;
  staticBackdrop: boolean = true;
  closable: boolean = false;

  createdComparator: Comparator<Tag> = new CustomComparator<Tag>('created', 'date');

  loading: boolean = false;

  @ViewChild('confirmationDialog')
  confirmationDialog: ConfirmationDialogComponent;

  constructor(
    private errorHandler: ErrorHandler,
    private tagService: TagService,
    private translateService: TranslateService,
    private ref: ChangeDetectorRef){}

  confirmDeletion(message: ConfirmationAcknowledgement) {
    if (message &&
        message.source === ConfirmationTargets.TAG
        && message.state === ConfirmationState.CONFIRMED) {
        let tag: Tag = message.data;
        if (tag) {
            if (tag.signature) {
                return;
            } else {
                let tagName = tag.name;
                toPromise<number>(this.tagService
                .deleteTag(this.repoName, tagName))
                .then(
                response => {
                    this.retrieve();
                    this.translateService.get('REPOSITORY.DELETED_TAG_SUCCESS')
                        .subscribe(res=>this.errorHandler.info(res));
                }).catch(error => this.errorHandler.error(error));
            }
        }
    }
  }

  ngOnInit() {
    if(!this.projectId) {
      this.errorHandler.error('Project ID cannot be unset.');
      return;
    }
    if(!this.repoName) {
      this.errorHandler.error('Repo name cannot be unset.');
      return;
    }
    if(!this.sessionInfo) {
      this.errorHandler.error('Session info cannot be unset.');
      return;
    }
    this.hasSignedIn = this.sessionInfo.hasSignedIn || false;
    this.hasProjectAdminRole = this.sessionInfo.hasProjectAdminRole || false;
    this.registryUrl = this.sessionInfo.registryUrl || '';
    this.withNotary = this.sessionInfo.withNotary || false;

    this.retrieve(); 
  }

  retrieve() {
    this.tags = [];
    this.loading = true;
    toPromise<Tag[]>(this.tagService
        .getTags(this.repoName))
        .then(items => { 
          this.tags = items;
          this.loading = false;
        })
        .catch(error => {
          this.errorHandler.error(error);
          this.loading = false;
        });
    let hnd = setInterval(()=>this.ref.markForCheck(), 100);
    setTimeout(()=>clearInterval(hnd), 1000);
  }

  deleteTag(tag: Tag) {
    if (tag) {
      let titleKey: string, summaryKey: string, content: string, buttons: ConfirmationButtons;
      if (tag.signature) {
        titleKey = 'REPOSITORY.DELETION_TITLE_TAG_DENIED';
        summaryKey = 'REPOSITORY.DELETION_SUMMARY_TAG_DENIED';
        buttons = ConfirmationButtons.CLOSE;
        content = 'notary -s https://' + this.registryUrl + ':4443 -d ~/.docker/trust remove -p ' + this.registryUrl + '/' + this.repoName + ' ' + tag.name;
      } else {
        titleKey = 'REPOSITORY.DELETION_TITLE_TAG';
        summaryKey = 'REPOSITORY.DELETION_SUMMARY_TAG';
        buttons = ConfirmationButtons.DELETE_CANCEL;
        content = tag.name;
      }
      let message = new ConfirmationMessage(
        titleKey,
        summaryKey,
        content,
        tag,
        ConfirmationTargets.TAG,
        buttons);
      this.confirmationDialog.open(message);
    }
  }

  showTagID(type: string, tag: Tag) {
    if(tag) {
      if(type === 'tag') {
        this.manifestInfoTitle = 'REPOSITORY.COPY_ID';
        this.tagID = tag.digest;
      } else if(type === 'parent') {
        this.manifestInfoTitle = 'REPOSITORY.COPY_PARENT_ID';
        this.tagID = tag.digest;
      }
      this.showTagManifestOpened = true;
    }
  }
  selectAndCopy($event: any) {
    $event.target.select();
  }
}