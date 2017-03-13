import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { RepositoryService } from '../repository.service';
import { MessageService } from '../../global-message/message.service';
import { AlertType, DeletionTargets } from '../../shared/shared.const';

import { DeletionDialogService } from '../../shared/deletion-dialog/deletion-dialog.service';
import { DeletionMessage } from '../../shared/deletion-dialog/deletion-message';

import { Subscription } from 'rxjs/Subscription';

import { TagView } from '../tag-view';

@Component({
  selector: 'tag-repository',
  templateUrl: 'tag-repository.component.html'
})
export class TagRepositoryComponent implements OnInit, OnDestroy {

  projectId: number;
  repoName: string;

  tags: TagView[];

  private subscription: Subscription;

  constructor(
    private route: ActivatedRoute,
    private messageService: MessageService,
    private deletionDialogService: DeletionDialogService,
    private repositoryService: RepositoryService) {
      this.subscription = this.deletionDialogService.deletionConfirm$.subscribe(
        message=>{
          let tag = message.data;
          if(tag) {
            if(tag.verified) {
              return;
            } else {
              let tagName = tag.tag;
              this.repositoryService
                  .deleteRepoByTag(this.repoName, tagName)
                  .subscribe(
                    response=>{
                      this.retrieve();
                      console.log('Deleted repo:' + this.repoName + ' with tag:' + tagName);
                    },
                    error=>this.messageService.announceMessage(error.status, 'Failed to delete tag:' + tagName + ' under repo:' + this.repoName, AlertType.DANGER)
                  );
            }
          }
          
        }
      )
  }
  
  ngOnInit() {
    this.projectId = this.route.snapshot.params['id'];  
    this.repoName = this.route.snapshot.params['repo'];
    this.tags = [];
    this.retrieve();
  }

  ngOnDestroy() {
    if(this.subscription) {
      this.subscription.unsubscribe();
    }
  }
  
  retrieve() {
    this.tags = [];
    this.repositoryService
        .listTagsWithVerifiedSignatures(this.repoName)
        .subscribe(
          items=>{
            items.forEach(t=>{
              let tag = new TagView();
              tag.tag = t.tag;
              let data = JSON.parse(t.manifest.history[0].v1Compatibility);
              tag.architecture = data['architecture'];
              tag.author = data['author'];
              tag.verified = t.verified || false;
              tag.created = data['created'];
              tag.dockerVersion = data['docker_version'];
              tag.pullCommand = 'docker pull ' + t.manifest.name + ':' + t.tag;
              tag.os = data['os'];
              this.tags.push(tag);
            });
        },
        error=>this.messageService.announceMessage(error.status, 'Failed to list tags with repo:' + this.repoName, AlertType.DANGER));
  }

  deleteTag(tag: TagView) {
    if(tag) {
      let titleKey: string, summaryKey: string;
      if (tag.verified) {
        titleKey = 'REPOSITORY.DELETION_TITLE_TAG_DENIED';
        summaryKey = 'REPOSITORY.DELETION_SUMMARY_TAG_DENIED';
      } else {
        titleKey = 'REPOSITORY.DELETION_TITLE_TAG';
        summaryKey = 'REPOSITORY.DELETION_SUMMARY_TAG';
      }
      let message = new DeletionMessage(titleKey, summaryKey, tag.tag, tag, DeletionTargets.TAG);
      this.deletionDialogService.openComfirmDialog(message);
    }
  }

}