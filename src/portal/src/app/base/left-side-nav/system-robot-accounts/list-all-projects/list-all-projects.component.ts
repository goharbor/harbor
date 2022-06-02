import { Component, Input } from '@angular/core';
import { Project } from '../../../../../../ng-swagger-gen/models/project';
import { clone, CustomComparator } from '../../../../shared/units/utils';
import { ClrDatagridComparatorInterface } from '@clr/angular';
import { Router } from '@angular/router';
import {
    ACTION_RESOURCE_I18N_MAP,
    FrontAccess,
    FrontProjectForAdd,
    INITIAL_ACCESSES,
    PermissionsKinds,
} from '../system-robot-util';
import { RobotPermission } from '../../../../../../ng-swagger-gen/models/robot-permission';

@Component({
    selector: 'app-list-all-projects',
    templateUrl: './list-all-projects.component.html',
    styleUrls: ['./list-all-projects.component.scss'],
})
export class ListAllProjectsComponent {
    cachedAllProjects: Project[];
    i18nMap = ACTION_RESOURCE_I18N_MAP;
    permissionsForAdd: RobotPermission[] = [];
    selectedRow: FrontProjectForAdd[] = [];
    timeComparator: ClrDatagridComparatorInterface<Project> =
        new CustomComparator<Project>('creation_time', 'date');
    projects: FrontProjectForAdd[] = [];
    pageSize: number = 5;
    currentPage: number = 1;
    defaultAccesses: FrontAccess[] = [];
    @Input()
    coverAll: boolean = false;
    showSelectAll: boolean = true;
    myNameFilterValue: string;
    constructor(private router: Router) {}

    init(isEdit: boolean) {
        this.pageSize = 5;
        this.currentPage = 1;
        this.showSelectAll = true;
        this.myNameFilterValue = null;
        if (isEdit) {
            this.defaultAccesses = clone(INITIAL_ACCESSES);
            this.defaultAccesses.forEach(item => (item.checked = false));
        } else {
            this.defaultAccesses = clone(INITIAL_ACCESSES);
        }
        if (this.cachedAllProjects && this.cachedAllProjects.length) {
            this.projects = clone(this.cachedAllProjects);
            this.resetAccess(this.defaultAccesses);
        } else {
            this.projects = [];
        }
    }
    resetAccess(accesses: FrontAccess[]) {
        if (this.projects && this.projects.length) {
            this.projects.forEach(item => {
                item.permissions = [
                    {
                        kind: PermissionsKinds.PROJECT,
                        namespace: item.name,
                        access: clone(accesses),
                    },
                ];
            });
        }
    }
    chooseAccess(access: FrontAccess) {
        access.checked = !access.checked;
    }
    chooseDefaultAccess(access: FrontAccess) {
        access.checked = !access.checked;
        this.resetAccess(this.defaultAccesses);
    }
    getPermissions(access: FrontAccess[]): number {
        let count: number = 0;
        access.forEach(item => {
            if (item.checked) {
                count++;
            }
        });
        return count;
    }
    getLink(proId: number): string {
        return `/harbor/projects/${proId}`;
    }
    selectAllOrUnselectAll() {
        if (this.showSelectAll) {
            if (this.myNameFilterValue) {
                this.projects.forEach(item => {
                    let flag = false;
                    if (item.name.indexOf(this.myNameFilterValue) !== -1) {
                        this.selectedRow.forEach(item2 => {
                            if (item2.name === item.name) {
                                flag = true;
                            }
                        });
                        if (!flag) {
                            this.selectedRow.push(item);
                        }
                    }
                });
            } else {
                this.selectedRow = this.projects;
            }
        } else {
            this.selectedRow = [];
        }
        this.showSelectAll = !this.showSelectAll;
    }
    isSelectAll(permissions: FrontAccess[]): boolean {
        if (permissions?.length) {
            return (
                permissions.filter(item => item.checked).length <
                permissions.length / 2
            );
        }
        return false;
    }
    selectAllPermissionOrUnselectAll(permissions: FrontAccess[]) {
        if (permissions?.length) {
            if (this.isSelectAll(permissions)) {
                permissions.forEach(item => {
                    item.checked = true;
                });
            } else {
                permissions.forEach(item => {
                    item.checked = false;
                });
            }
        }
    }
}
