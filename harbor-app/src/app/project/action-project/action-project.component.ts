import { Component, EventEmitter, Input, Output } from '@angular/core';
import { Project } from '../project';
import { ProjectService } from '../project.service';

import { TranslateService } from '@ngx-translate/core';
import { DeletionDialogService } from '../../shared/deletion-dialog/deletion-dialog.service';
import { DeletionMessage } from '../../shared/deletion-dialog/deletion-message';
import { DeletionTargets } from '../../shared/shared.const';

@Component({
  selector: 'action-project',
  templateUrl: 'action-project.component.html'
})
export class ActionProjectComponent {

  @Output() togglePublic = new EventEmitter<Project>();
  @Output() deleteProject = new EventEmitter<Project>();

  @Input() project: Project;

  constructor(private projectService: ProjectService,
    private deletionDialogService: DeletionDialogService,
    private translateService: TranslateService) {
    deletionDialogService.deletionConfirm$.subscribe(message => {
      if (message && message.targetId === DeletionTargets.PROJECT) {
        this.deleteProject.emit(message.data);
      }
    });
  }

  toggle() {
    if (this.project) {
      this.project.public === 0 ? this.project.public = 1 : this.project.public = 0;
      this.togglePublic.emit(this.project);
    }
  }

  delete() {
    // if(this.project) {
    //   this.deleteProject.emit(this.project);
    // }
    let deletionMessage = new DeletionMessage(
      'PROJECT.DELETION_TITLE',
      'PROJECT.DELETION_SUMMARY',
      this.project.name,
      this.project,
      DeletionTargets.PROJECT
    );
    this.deletionDialogService.openComfirmDialog(deletionMessage);
  }
}