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
import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { RepositoryService } from './repository.service';
import { Repository } from './repository';

import { MessageHandlerService } from '../shared/message-handler/message-handler.service';
import { ConfirmationState, ConfirmationTargets } from '../shared/shared.const';


import { ConfirmationDialogService } from '../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../shared/confirmation-dialog/confirmation-message';
import { Subscription } from 'rxjs/Subscription';

import { State } from 'clarity-angular';

import { Project } from '../project/project';

@Component({
  selector: 'repository',
  templateUrl: 'repository.component.html',
  styleUrls: ['./repository.component.css']
})
export class RepositoryComponent implements OnInit {
  changedRepositories: Repository[];

  projectId: number;

  lastFilteredRepoName: string;

  page: number = 1;
  pageSize: number = 15;

  totalPage: number;
  totalRecordCount: number;

  hasProjectAdminRole: boolean;

  subscription: Subscription;

  constructor(
    private route: ActivatedRoute,
    private repositoryService: RepositoryService,
    private messageHandlerService: MessageHandlerService,
    private deletionDialogService: ConfirmationDialogService
  ) {
    this.subscription = this.deletionDialogService
      .confirmationConfirm$
      .subscribe(
      message => {
        if (message &&
          message.source === ConfirmationTargets.REPOSITORY &&
          message.state === ConfirmationState.CONFIRMED) {
          let repoName = message.data;
          this.repositoryService
            .deleteRepository(repoName)
            .subscribe(
            response => {
              this.refresh();
              this.messageHandlerService.showSuccess('REPOSITORY.DELETED_REPO_SUCCESS');
              console.log('Successful deleted repo:' + repoName);
            },
            error => this.messageHandlerService.handleError(error)
          );
        }
      });
 
  }

  ngOnInit(): void {
    this.projectId = this.route.snapshot.parent.params['id'];
    let resolverData = this.route.snapshot.parent.data;
    if(resolverData) {
      this.hasProjectAdminRole = (<Project>resolverData['projectResolver']).has_project_admin_role;
    }
    this.lastFilteredRepoName = '';
    this.retrieve();
  }

  ngOnDestroy(): void {
    if (this.subscription) {
      this.subscription.unsubscribe();
    }
  }

  retrieve(state?: State) {
    if (state) {
      this.page = state.page.to + 1;
    }
    this.repositoryService
      .listRepositories(this.projectId, this.lastFilteredRepoName, this.page, this.pageSize)
      .subscribe(
      response => {
        this.totalRecordCount = response.headers.get('x-total-count');
        this.totalPage = Math.ceil(this.totalRecordCount / this.pageSize);
        console.log('TotalRecordCount:' + this.totalRecordCount + ', totalPage:' + this.totalPage);
        this.changedRepositories = response.json();
      },
      error => this.messageHandlerService.handleError(error)
      );
  }

  doSearchRepoNames(repoName: string) {
    this.lastFilteredRepoName = repoName;
    this.retrieve();

  }

  deleteRepo(repoName: string) {
    let message = new ConfirmationMessage(
      'REPOSITORY.DELETION_TITLE_REPO',
      'REPOSITORY.DELETION_SUMMARY_REPO',
      repoName,
      repoName,
      ConfirmationTargets.REPOSITORY);
    this.deletionDialogService.openComfirmDialog(message);
  }

  refresh() {
    this.retrieve();
  }
}