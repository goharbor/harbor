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
import { ConfirmationTargets, ConfirmationState } from '../../shared/shared.const';

import { ConfirmationDialogService } from '../../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../../shared/confirmation-dialog/confirmation-message';

import { Subscription } from 'rxjs/Subscription';

import { Tag } from '../tag';
import { TagView } from '../tag-view';

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

  tags: TagView[];
  registryUrl: string;
  withNotary: boolean;

  hasSignedIn: boolean;

  showTagManifestOpened: boolean;
  manifestInfoTitle: string;
  tagID: string;
  staticBackdrop: boolean = true;
  closable: boolean = false;

  selectAll: boolean = false;

  private subscription: Subscription;

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
              let tagName = tag.tag;
              this.repositoryService
                .deleteRepoByTag(this.repoName, tagName)
                .subscribe(
                response => {
                  this.retrieve();
                  this.messageHandlerService.showSuccess('REPOSITORY.DELETED_TAG_SUCCESS');
                  console.log('Deleted repo:' + this.repoName + ' with tag:' + tagName);
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
    this.tags = [];
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
          items => this.listTags(items),
          error => this.messageHandlerService.handleError(error));
   
    if(this.withNotary) {
      this.repositoryService
          .listNotarySignatures(this.repoName)
          .subscribe(
            signatures => {
              this.tags.forEach((t, n)=>{
                let signed = false;
                for(let i = 0; i < signatures.length; i++) {
                  if (signatures[i].tag === t.tag) {
                    signed = true;
                    break;
                  }
                }
                this.tags[n].signed = (signed) ? 1 : 0;
                this.ref.markForCheck();
              });
            },
            error => console.error('Cannot determine the signature of this tag.'));
      }
  }

  private listTags(tags: Tag[]): void {
    tags.forEach(t => {
      let tag = new TagView();
      tag.tag = t.tag;
      let data = JSON.parse(t.manifest.history[0].v1Compatibility);
      tag.architecture = data['architecture'];
      tag.author = data['author'];
      tag.signed = t.signed;
      tag.created = data['created'];
      tag.dockerVersion = data['docker_version'];
      tag.pullCommand = 'docker pull ' + this.registryUrl + '/' + t.manifest.name + ':' + t.tag;
      tag.os = data['os'];
      tag.id = data['id'];
      tag.parent = data['parent'];
      this.tags.push(tag);
    });
    let hnd = setInterval(()=>this.ref.markForCheck(), 100);
    setTimeout(()=>clearInterval(hnd), 1000);
  }

  deleteTag(tag: TagView) {
    if (tag) {
      let titleKey: string, summaryKey: string, content: string, confirmOnly: boolean;
      if (tag.signed) {
        titleKey = 'REPOSITORY.DELETION_TITLE_TAG_DENIED';
        summaryKey = 'REPOSITORY.DELETION_SUMMARY_TAG_DENIED';
        confirmOnly = true;
        content = 'notary -s https://' + this.registryUrl + ':4443 -d ~/.docker/trust remove -p ' + this.registryUrl + '/' + this.repoName + ' ' + tag.tag;
      } else {
        titleKey = 'REPOSITORY.DELETION_TITLE_TAG';
        summaryKey = 'REPOSITORY.DELETION_SUMMARY_TAG';
        content = tag.tag;
        confirmOnly = false;
      }
      let message = new ConfirmationMessage(
        titleKey,
        summaryKey,
        content,
        tag,
        ConfirmationTargets.TAG);
        message.confirmOnly = confirmOnly;
      this.deletionDialogService.openComfirmDialog(message);
    }
  }

  showTagID(type: string, tag: TagView) {
    if(tag) {
      if(type === 'tag') {
        this.manifestInfoTitle = 'REPOSITORY.COPY_ID';
        this.tagID = tag.id;
      } else if(type === 'parent') {
        this.manifestInfoTitle = 'REPOSITORY.COPY_PARENT_ID';
        this.tagID = tag.parent;
      }
      this.showTagManifestOpened = true;
    }
  }
  selectAndCopy($event) {
    $event.target.select();
  }
}