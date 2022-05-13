import { Project } from '../../project';
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

import { SessionService } from '../../../../shared/services/session.service';
import { SessionUser } from '../../../../shared/entities/session-user';

@Component({
    selector: 'project-list-charts',
    templateUrl: './list-charts.component.html',
    styleUrls: ['./list-charts.component.scss'],
})
export class ListChartsComponent implements OnInit {
    projectId: number;

    projectName: string;
    urlPrefix: string;
    hasSignedIn: boolean;
    project_member_role_id: number;
    currentUser: SessionUser;

    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private session: SessionService
    ) {}

    ngOnInit() {
        // Get projectId from router-guard params snapshot.
        this.projectId = +this.route.snapshot.parent.parent.params['id'];
        // Get current user from registered resolver.
        this.currentUser = this.session.getCurrentUser();
        let resolverData = this.route.snapshot.parent.parent.data;
        if (resolverData) {
            let project = <Project>resolverData['projectResolver'];
            this.projectName = project.name;
            this.project_member_role_id = project.current_user_role_id;
        }
    }

    onChartClick(chartName: string) {
        this.router.navigateByUrl(`${this.router.url}/${chartName}/versions`);
    }
}
