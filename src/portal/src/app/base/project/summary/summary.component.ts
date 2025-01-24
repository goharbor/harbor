import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { AppConfigService } from '../../../services/app-config.service';
import {
    Endpoint,
    ProjectService,
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../shared/services';
import { ErrorHandler } from '../../../shared/units/error-handler';
import {
    FALSE_STR,
    PROJECT_SUMMARY_CARD_VIEW_LOCALSTORAGE_KEY,
    TRUE_STR,
} from '../../../shared/entities/shared.const';
import { RepositoryService } from '../../../../../ng-swagger-gen/services/repository.service';
import { Project } from '../../../../../ng-swagger-gen/models/project';
import { Repository } from '../../../../../ng-swagger-gen/models/repository';

@Component({
    selector: 'summary',
    templateUrl: './summary.component.html',
    styleUrls: ['./summary.component.scss'],
})
export class SummaryComponent implements OnInit {
    showProjectMemberInfo: boolean = false;
    hasReadRepoPermission: boolean = false;
    projectId: number;
    projectName: string;
    summaryInformation: any;
    endpoint: Endpoint;
    isCardView: boolean = true;
    cardHover: boolean = false;
    listHover: boolean = false;
    repos: Repository[] = [];
    constructor(
        private projectService: ProjectService,
        private userPermissionService: UserPermissionService,
        private errorHandler: ErrorHandler,
        private route: ActivatedRoute,
        private repoService: RepositoryService,
        private router: Router
    ) {
        if (localStorage) {
            if (
                !localStorage.getItem(
                    PROJECT_SUMMARY_CARD_VIEW_LOCALSTORAGE_KEY
                )
            ) {
                localStorage.setItem(
                    PROJECT_SUMMARY_CARD_VIEW_LOCALSTORAGE_KEY,
                    FALSE_STR
                );
            }
            this.isCardView =
                localStorage.getItem(
                    PROJECT_SUMMARY_CARD_VIEW_LOCALSTORAGE_KEY
                ) === TRUE_STR;
        }
    }

    ngOnInit() {
        this.projectId = this.route.parent.parent.snapshot.params['id'];
        const resolverData = this.route.snapshot.parent.parent.data;
        if (resolverData) {
            let project = <Project>resolverData['projectResolver'];
            this.projectName = project.name;
        }
        const permissions = [
            {
                resource: USERSTATICPERMISSION.MEMBER.KEY,
                action: USERSTATICPERMISSION.MEMBER.VALUE.LIST,
            },
            {
                resource: USERSTATICPERMISSION.REPOSITORY.KEY,
                action: USERSTATICPERMISSION.REPOSITORY.VALUE.LIST,
            },
        ];
        this.userPermissionService
            .hasProjectPermissions(this.projectId, permissions)
            .subscribe((results: Array<boolean>) => {
                this.showProjectMemberInfo = results[0];
                this.hasReadRepoPermission = results[1];
            });
        this.projectService.getProjectSummary(this.projectId).subscribe(
            res => {
                this.summaryInformation = res;
            },
            error => {
                this.errorHandler.error(error);
            }
        );
        if (this.isCardView) {
            this.getDataForCardView();
        }
    }
    showCard(cardView: boolean) {
        if (this.isCardView === cardView) {
            return;
        }
        this.isCardView = cardView;
        if (localStorage) {
            if (this.isCardView) {
                localStorage.setItem(
                    PROJECT_SUMMARY_CARD_VIEW_LOCALSTORAGE_KEY,
                    TRUE_STR
                );
            } else {
                localStorage.setItem(
                    PROJECT_SUMMARY_CARD_VIEW_LOCALSTORAGE_KEY,
                    FALSE_STR
                );
            }
        }
        if (this.isCardView) {
            this.getDataForCardView();
        }
    }

    mouseEnter(itemName: string) {
        if (itemName === 'card') {
            this.cardHover = true;
        } else {
            this.listHover = true;
        }
    }

    mouseLeave(itemName: string) {
        if (itemName === 'card') {
            this.cardHover = false;
        } else {
            this.listHover = false;
        }
    }

    isHovering(itemName: string) {
        if (itemName === 'card') {
            return this.cardHover;
        } else {
            return this.listHover;
        }
    }
    getDataForCardView() {
        this.getTop4Repos();
    }
    getTop4Repos() {
        if (this.hasReadRepoPermission) {
            this.repoService
                .listRepositories({
                    projectName: this.projectName,
                    page: 1,
                    pageSize: 4,
                })
                .subscribe(res => {
                    this.repos = res;
                });
        }
    }
    goIntoRepo(repoEvt: Repository): void {
        const linkUrl = [
            'harbor',
            'projects',
            repoEvt.project_id,
            'repositories',
            repoEvt.name.substr(this.projectName.length + 1),
        ];
        this.router.navigate(linkUrl);
    }
    goToRepos() {
        const linkUrl = ['harbor', 'projects', this.projectId, 'repositories'];
        this.router.navigate(linkUrl);
    }
    goToMembers() {
        const linkUrl = ['harbor', 'projects', this.projectId, 'members'];
        this.router.navigate(linkUrl);
    }
    getTotalMembers(): number {
        if (this.summaryInformation) {
            return (
                +(this.summaryInformation.project_admin_count
                    ? this.summaryInformation.project_admin_count
                    : 0) +
                +(this.summaryInformation.maintainer_count
                    ? this.summaryInformation.maintainer_count
                    : 0) +
                +(this.summaryInformation.developer_count
                    ? this.summaryInformation.developer_count
                    : 0) +
                +(this.summaryInformation.guest_count
                    ? this.summaryInformation.guest_count
                    : 0) +
                +(this.summaryInformation.limited_guest_count
                    ? this.summaryInformation.limited_guest_count
                    : 0)
            );
        }
        return 0;
    }
}
