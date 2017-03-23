import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { RepositoryService } from '../repository.service';
import { MessageService } from '../../global-message/message.service';
import { AlertType, ConfirmationTargets, ConfirmationState } from '../../shared/shared.const';

import { ConfirmationDialogService } from '../../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../../shared/confirmation-dialog/confirmation-message';

import { Subscription } from 'rxjs/Subscription';

import { Tag } from '../tag';
import { TagView } from '../tag-view';

import { AppConfigService } from '../../app-config.service';

import { SessionService } from '../../shared/session.service';

import { Project } from '../../project/project';

@Component({
  moduleId: module.id,
  selector: 'tag-repository',
  templateUrl: 'tag-repository.component.html',
  styleUrls: ['./tag-repository.component.css']
})
export class TagRepositoryComponent implements OnInit, OnDestroy {

  projectId: number;
  repoName: string;

  hasProjectAdminRole: boolean = false;

  tags: TagView[];
  registryUrl: string;
  withNotary: boolean;

  private subscription: Subscription;

  constructor(
    private route: ActivatedRoute,
    private messageService: MessageService,
    private deletionDialogService: ConfirmationDialogService,
    private repositoryService: RepositoryService,
    private appConfigService: AppConfigService,
    private session: SessionService){
    
    this.subscription = this.deletionDialogService.confirmationConfirm$.subscribe(
      message => {
        if (message &&
          message.source === ConfirmationTargets.TAG
          && message.state === ConfirmationState.CONFIRMED) {
          let tag = message.data;
          if (tag) {
            if (tag.verified) {
              return;
            } else {
              let tagName = tag.tag;
              this.repositoryService
                .deleteRepoByTag(this.repoName, tagName)
                .subscribe(
                response => {
                  this.retrieve();
                  this.messageService.announceMessage(response, 'REPOSITORY.DELETED_TAG_SUCCESS', AlertType.SUCCESS);
                  console.log('Deleted repo:' + this.repoName + ' with tag:' + tagName);
                },
                error => this.messageService.announceMessage(error.status, 'Failed to delete tag:' + tagName + ' under repo:' + this.repoName, AlertType.DANGER)
                );
            }
          }
        }
      });
  }

  ngOnInit() {
    let resolverData = this.route.snapshot.data;
    console.log(JSON.stringify(resolverData));
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
    if(this.withNotary) {
      this.repositoryService
          .listTagsWithVerifiedSignatures(this.repoName)
          .subscribe(
            items => this.listTags(items),
            error => this.messageService.announceMessage(error.status, 'Failed to list tags with repo:' + this.repoName, AlertType.DANGER));
    } else {
      this.repositoryService
          .listTags(this.repoName)
          .subscribe(
            items => this.listTags(items),
            error => this.messageService.announceMessage(error.status, 'Failed to list tags with repo:' + this.repoName, AlertType.DANGER));
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
      this.tags.push(tag);
    });
  }

  deleteTag(tag: TagView) {
    if (tag) {
      let titleKey: string, summaryKey: string, content: string;
      if (tag.signed) {
        titleKey = 'REPOSITORY.DELETION_TITLE_TAG_DENIED';
        summaryKey = 'REPOSITORY.DELETION_SUMMARY_TAG_DENIED';
        content = 'notary -s https://' + this.registryUrl + ' -d ~/.docker/trust remove -p ' + this.registryUrl + '/' + this.repoName + ':' + tag.tag;
      } else {
        titleKey = 'REPOSITORY.DELETION_TITLE_TAG';
        summaryKey = 'REPOSITORY.DELETION_SUMMARY_TAG';
        content = tag.tag;
      }
      let message = new ConfirmationMessage(
        titleKey,
        summaryKey,
        content,
        tag,
        ConfirmationTargets.TAG);
      this.deletionDialogService.openComfirmDialog(message);
    }
  }

}