import { Component, EventEmitter, Input, Output } from '@angular/core';
import { Project } from '../project';
import { ProjectService } from '../project.service';

import { TranslateService } from '@ngx-translate/core';
import { DeletionDialogService } from '../../shared/deletion-dialog/deletion-dialog.service';
import { DeletionMessage } from '../../shared/deletion-dialog/deletion-message';

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
    deletionDialogService.deletionConfirm$.subscribe(project=>this.deleteProject.emit(project));
  }

  toggle() {
    if(this.project) {
      this.project.public === 0 ? this.project.public = 1 : this.project.public = 0;
      this.togglePublic.emit(this.project);
    }
  }

  delete() {
    // if(this.project) {
    //   this.deleteProject.emit(this.project);
    // }
    let deletionMessage = new DeletionMessage('Delete Project', 'Do you confirm to delete project?', this.project);
    this.deletionDialogService.openComfirmDialog(deletionMessage);
  }
}