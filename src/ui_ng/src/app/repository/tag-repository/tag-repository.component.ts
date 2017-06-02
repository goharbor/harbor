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
import { Component, OnInit, OnDestroy, ChangeDetectionStrategy, ChangeDetectorRef } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { RepositoryService } from '../repository.service';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { ConfirmationTargets, ConfirmationState, ConfirmationButtons } from '../../shared/shared.const';

import { ConfirmationDialogService } from '../../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../../shared/confirmation-dialog/confirmation-message';

import { Subscription } from 'rxjs/Subscription';

import { Tag } from '../tag';

import { AppConfigService } from '../../app-config.service';

import { SessionService } from '../../shared/session.service';

import { Project } from '../../project/project';

@Component({
  selector: 'tag-repository',
  templateUrl: 'tag-repository.component.html',
  styleUrls: ['./tag-repository.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class TagRepositoryComponent implements OnInit, OnDestroy {

  projectId: number;
  repoName: string;

  hasProjectAdminRole: boolean = false;

  tags: Tag[];
  registryUrl: string;
  withNotary: boolean;

  hasSignedIn: boolean;

  showTagManifestOpened: boolean;
  manifestInfoTitle: string;
  digestId: string;
  staticBackdrop: boolean = true;
  closable: boolean = false;

  selectAll: boolean = false;

  subscription: Subscription;

  constructor(
    private route: ActivatedRoute,
    private messageHandlerService: MessageHandlerService,
    private deletionDialogService: ConfirmationDialogService,
    private repositoryService: RepositoryService,
    private appConfigService: AppConfigService,
    private session: SessionService,
    private ref: ChangeDetectorRef){
    this.subscription = this.deletionDialogService.confirmationConfirm$.subscribe(
      message => {
        if (message &&
          message.source === ConfirmationTargets.TAG
          && message.state === ConfirmationState.CONFIRMED) {
          let tag = message.data;
          if (tag) {
            if (tag.signed) {
              return;
            } else {
              this.repositoryService
                .deleteRepoByTag(this.repoName, tag.name)
                .subscribe(
                response => {
                  this.retrieve();
                  this.messageHandlerService.showSuccess('REPOSITORY.DELETED_TAG_SUCCESS');
                },
                error => this.messageHandlerService.handleError(error)
              );
            }
          }
        }
      });
  }

  ngOnInit() {
    this.hasSignedIn = (this.session.getCurrentUser() !== null);
    let resolverData = this.route.snapshot.data;
    if(resolverData) {
      this.hasProjectAdminRole = (<Project>resolverData['projectResolver']).has_project_admin_role;
    }
    this.projectId = this.route.snapshot.params['id'];
    this.repoName = this.route.snapshot.params['repo'];

    this.registryUrl = this.appConfigService.getConfig().registry_url;
    this.withNotary = this.appConfigService.getConfig().with_notary;
    this.retrieve();
    
  }

  ngOnDestroy() {
    if (this.subscription) {
      this.subscription.unsubscribe();
    }
  }

  retrieve() {
    this.tags = [];
    this.repositoryService
        .listTags(this.repoName)
        .subscribe(
          tags => this.tags = tags,
          error => this.messageHandlerService.handleError(error));
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
      this.deletionDialogService.openComfirmDialog(message);
    }
  }

  showDigestId(tag: Tag) {
    if(tag) {
      this.manifestInfoTitle = 'REPOSITORY.COPY_DIGEST_ID';
      this.digestId = tag.digest;
      this.showTagManifestOpened = true;
    }
  }
  selectAndCopy($event: any) {
    $event.target.select();
  }
}