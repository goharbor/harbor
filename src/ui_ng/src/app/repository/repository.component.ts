import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { RepositoryService } from './repository.service';
import { Repository } from './repository';

import { MessageService } from '../global-message/message.service';
import { AlertType, DeletionTargets } from '../shared/shared.const';


import { DeletionDialogService } from '../shared/deletion-dialog/deletion-dialog.service';
import { DeletionMessage } from '../shared/deletion-dialog/deletion-message';
import { Subscription } from 'rxjs/Subscription';

const repositoryTypes = [
  { key: '0', description: 'REPOSITORY.MY_REPOSITORY' },
  { key: '1', description: 'REPOSITORY.PUBLIC_REPOSITORY' }
];

@Component({
  selector: 'repository',
  templateUrl: 'repository.component.html'
})
export class RepositoryComponent implements OnInit {
  changedRepositories: Repository[];

  projectId: number;
  repositoryTypes = repositoryTypes;
  currentRepositoryType: {};
  lastFilteredRepoName: string;

  subscription: Subscription;

  constructor(
    private route: ActivatedRoute,
    private repositoryService: RepositoryService,
    private messageService: MessageService,
    private deletionDialogService: DeletionDialogService
  ) {
    this.subscription = this.deletionDialogService
        .deletionConfirm$
        .subscribe(
          message=>{
            let repoName = message.data;
            this.repositoryService
                .deleteRepository(repoName)
                .subscribe(
                  response=>{
                    this.refresh();
                    console.log('Successful deleted repo:' + repoName);
                  },
                  error=>this.messageService.announceMessage(error.status, 'Failed to delete repo:' + repoName, AlertType.DANGER)
                );
          }
        );
  }

  ngOnInit(): void {
    this.projectId = this.route.snapshot.parent.params['id'];
    this.currentRepositoryType = this.repositoryTypes[0];
    this.lastFilteredRepoName = '';
    this.retrieve(this.lastFilteredRepoName);
  }

  ngOnDestroy(): void {
    if(this.subscription) {
      this.subscription.unsubscribe();
    }
  }

  retrieve(repoName: string) {
    this.repositoryService
        .listRepositories(this.projectId, repoName)
        .subscribe(
          response=>this.changedRepositories=response,
          error=>this.messageService.announceMessage(error.status, 'Failed to list repositories.', AlertType.DANGER)
        );
  }

  doFilterRepositoryByType(type: string) {
    this.currentRepositoryType = this.repositoryTypes.find(r=>r.key == type);
  }
  
  doSearchRepoNames(repoName: string) {
    this.lastFilteredRepoName = repoName;
    this.retrieve(this.lastFilteredRepoName);
   
  }

  deleteRepo(repoName: string) {
    let message = new DeletionMessage(
      'REPOSITORY.DELETION_TITLE_REPO', 
      'REPOSITORY.DELETION_SUMMARY_REPO', 
      repoName, repoName, DeletionTargets.REPOSITORY);
    this.deletionDialogService.openComfirmDialog(message);
  }

  refresh() {
    this.retrieve('');
  }
}