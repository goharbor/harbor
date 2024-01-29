import { ChangeDetectorRef, Component, Input, OnInit } from '@angular/core';
import { Project } from '../../../../../../ng-swagger-gen/models/project';
import { clone, CustomComparator } from '../../../../shared/units/utils';
import { ClrDatagridComparatorInterface } from '@clr/angular';
import { FrontAccess } from '../system-robot-util';
import { Access } from '../../../../../../ng-swagger-gen/models/access';
import { forkJoin, Observable } from 'rxjs';
import { finalize } from 'rxjs/operators';
import { ProjectService } from '../../../../../../ng-swagger-gen/services/project.service';
import { PermissionSelectPanelModes } from '../../../../shared/components/robot-permissions-panel/robot-permissions-panel.component';
import { RobotPermission } from '../../../../../../ng-swagger-gen/models/robot-permission';
import { Permissions } from '../../../../../../ng-swagger-gen/models/permissions';

const FIRST_PROJECTS_PAGE_SIZE: number = 100;

@Component({
    selector: 'app-list-all-projects',
    templateUrl: './list-all-projects.component.html',
    styleUrls: ['./list-all-projects.component.scss'],
})
export class ListAllProjectsComponent implements OnInit {
    selectedRow: Project[] = [];
    selectedRowForEdit: Project[] = [];
    timeComparator: ClrDatagridComparatorInterface<Project> =
        new CustomComparator<Project>('creation_time', 'date');
    projects: Project[] = [];
    pageSize: number = 5;
    currentPage: number = 1;
    showSelectAll: boolean = true;
    myNameFilterValue: string;

    @Input()
    robotMetadata: Permissions;

    initialAccess: Access[] = [];
    selectedProjectPermissionMap: { [key: string]: Access[] } = {};
    selectedProjectPermissionMapForEdit: { [key: string]: Access[] } = {};
    loadingData: boolean = true;
    @Input()
    initDataForEdit: RobotPermission[];

    constructor(
        private projectService: ProjectService,
        private cdf: ChangeDetectorRef
    ) {}

    ngOnInit() {
        this.loadDataFromBackend();
    }
    resetAccess(accesses: FrontAccess[]) {
        if (this.projects && this.projects.length) {
            this.projects.forEach(item => {
                this.selectedProjectPermissionMap[item.name] = clone(accesses);
            });
        }
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

    loadDataFromBackend() {
        this.loadingData = true;
        this.projectService
            .listProjectsResponse({
                withDetail: false,
                page: 1,
                pageSize: FIRST_PROJECTS_PAGE_SIZE,
            })
            .subscribe({
                next: result => {
                    // Get total count
                    if (result.headers) {
                        const xHeader: string =
                            result.headers.get('X-Total-Count');
                        const totalCount = parseInt(xHeader, 0);
                        let arr = result.body || [];
                        if (totalCount <= FIRST_PROJECTS_PAGE_SIZE) {
                            // already gotten all projects
                            this.projects = result.body;
                            this.initDataForEditMode();
                            this.loadingData = false;
                            this.cdf.detectChanges();
                        } else {
                            // get all the projects in specified times
                            const times: number = Math.ceil(
                                totalCount / FIRST_PROJECTS_PAGE_SIZE
                            );
                            const observableList: Observable<Project[]>[] = [];
                            for (let i = 2; i <= times; i++) {
                                observableList.push(
                                    this.projectService.listProjects({
                                        withDetail: false,
                                        page: i,
                                        pageSize: FIRST_PROJECTS_PAGE_SIZE,
                                    })
                                );
                            }
                            forkJoin(observableList)
                                .pipe(
                                    finalize(() => {
                                        this.loadingData = false;
                                    })
                                )
                                .subscribe(res => {
                                    if (res && res.length) {
                                        res.forEach(item => {
                                            arr = arr.concat(item);
                                        });
                                        this.projects = arr;
                                        this.initDataForEditMode();
                                        this.cdf.detectChanges();
                                    }
                                });
                        }
                    }
                },
                error: error => {
                    this.loadingData = false;
                },
            });
    }
    initDataForEditMode() {
        if (this.initDataForEdit?.length) {
            this.selectedRow = [];
            this.projects.forEach((pro, index) => {
                this.initDataForEdit.forEach(item => {
                    if (pro.name === item.namespace) {
                        item.access.forEach(acc => {
                            this.selectedProjectPermissionMap[pro.name] =
                                item.access;
                        });
                        this.selectedRow.push(pro);
                    }
                });
                this.selectedProjectPermissionMapForEdit = clone(
                    this.selectedProjectPermissionMap
                );
                this.selectedRowForEdit = clone(this.selectedRow);
            });
        }
    }
    protected readonly PermissionSelectPanelModes = PermissionSelectPanelModes;
}
