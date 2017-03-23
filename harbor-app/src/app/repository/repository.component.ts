import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { RepositoryService } from './repository.service';
import { Repository } from './repository';

import { MessageService } from '../global-message/message.service';
import { AlertType, ConfirmationState, ConfirmationTargets } from '../shared/shared.const';


import { ConfirmationDialogService } from '../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../shared/confirmation-dialog/confirmation-message';
import { Subscription } from 'rxjs/Subscription';

import { State } from 'clarity-angular';

const repositoryTypes = [
  { key: '0', description: 'REPOSITORY.MY_REPOSITORY' },
  { key: '1', description: 'REPOSITORY.PUBLIC_REPOSITORY' }
];

@Component({
  moduleId: module.id,
  selector: 'repository',
  templateUrl: 'repository.component.html',
  styleUrls: ['./repository.component.css']
})
export class RepositoryComponent implements OnInit {
  changedRepositories: Repository[];

  projectId: number;
  repositoryTypes = repositoryTypes;
  currentRepositoryType: {};
  lastFilteredRepoName: string;

  page: number = 1;
  pageSize: number = 15;

  totalPage: number;
  totalRecordCount: number;

  subscription: Subscription;

  constructor(
    private route: ActivatedRoute,
    private repositoryService: RepositoryService,
    private messageService: MessageService,
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
              this.messageService.announceMessage(response, 'REPOSITORY.DELETED_REPO_SUCCESS', AlertType.SUCCESS);
              console.log('Successful deleted repo:' + repoName);
            },
            error => this.messageService.announceMessage(error.status, 'Failed to delete repo:' + repoName, AlertType.DANGER)
            );
        }
      }
      );
  }

  ngOnInit(): void {
    this.projectId = this.route.snapshot.parent.params['id'];
    this.currentRepositoryType = this.repositoryTypes[0];
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
      error => this.messageService.announceMessage(error.status, 'Failed to list repositories.', AlertType.DANGER)
      );
  }

  doFilterRepositoryByType(type: string) {
    this.currentRepositoryType = this.repositoryTypes.find(r => r.key == type);
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