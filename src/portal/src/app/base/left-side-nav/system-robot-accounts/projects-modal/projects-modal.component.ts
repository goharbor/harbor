import { Component } from '@angular/core';
import { Project } from '../../../../../../ng-swagger-gen/models/project';
import { Router } from '@angular/router';
import { ACTION_RESOURCE_I18N_MAP } from '../system-robot-util';
import { RobotPermission } from '../../../../../../ng-swagger-gen/models/robot-permission';

@Component({
    selector: 'app-projects-modal',
    templateUrl: './projects-modal.component.html',
    styleUrls: ['./projects-modal.component.scss'],
})
export class ProjectsModalComponent {
    projectsModalOpened: boolean = false;
    robotName: string;
    cachedAllProjects: Project[];
    permissions: RobotPermission[] = [];
    i18nMap = ACTION_RESOURCE_I18N_MAP;
    constructor(private router: Router) {}

    close() {
        this.projectsModalOpened = false;
    }
    getProject(p: RobotPermission): Project {
        if (this.cachedAllProjects && this.cachedAllProjects.length) {
            for (let i = 0; i < this.cachedAllProjects.length; i++) {
                if (p.namespace === this.cachedAllProjects[i].name) {
                    return this.cachedAllProjects[i];
                }
            }
        }
        return null;
    }
    goToLink(proId: number): void {
        this.router.navigate(['harbor', 'projects', proId]);
    }
}
