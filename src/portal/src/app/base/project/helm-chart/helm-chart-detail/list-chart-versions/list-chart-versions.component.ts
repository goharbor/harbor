import { ActivatedRoute, Router } from '@angular/router';
import { Component, OnInit } from '@angular/core';

import { Project } from '../../../project';
import { SessionUser } from '../../../../../shared/entities/session-user';
import { SessionService } from '../../../../../shared/services/session.service';

@Component({
    selector: 'list-chart-version',
    templateUrl: './list-chart-versions.component.html',
    styleUrls: ['./list-chart-versions.component.scss'],
})
export class ListChartVersionsComponent implements OnInit {
    loading = false;

    projectId: number;
    projectName: string;
    chartName: string;
    roleName: string;

    hasSignedIn: boolean;
    currentUser: SessionUser;

    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private session: SessionService
    ) {}

    ngOnInit() {
        // Get projectId from router-guard params snapshot.
        this.projectId = +this.route.snapshot.parent.params['id'];
        this.chartName = this.route.snapshot.params['chart'];
        // Get current user from registered resolver.
        this.currentUser = this.session.getCurrentUser();
        let resolverData = this.route.snapshot.parent.data;
        if (resolverData) {
            let project = <Project>resolverData['projectResolver'];
            this.roleName = project.role_name;
            this.projectName = project.name;
        }
    }

    onVersionClick(version: string) {
        this.router.navigateByUrl(`${this.router.url}/${version}`);
    }

    gotoProjectList() {
        this.router.navigateByUrl('/harbor/projects');
    }

    gotoChartList() {
        this.router.navigateByUrl(
            `/harbor/projects/${this.projectId}/helm-charts`
        );
    }
}
