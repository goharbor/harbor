// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { Component, OnInit, ViewChild, AfterViewInit } from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';

import { ReplicationComponent } from 'harbor-ui';

import {SessionService} from "../shared/session.service";
import {Project} from "../project/project";
import {ProjectService} from "../project/project.service";

@Component({
  selector: 'replication',
  templateUrl: 'replication-page.component.html'
})
export class ReplicationPageComponent implements OnInit, AfterViewInit {
  projectIdentify: string | number;
  @ViewChild("replicationView") replicationView: ReplicationComponent;
  projectName: string;

  constructor(private route: ActivatedRoute,
              private router: Router,
              private proService: ProjectService,
              private session: SessionService) { }

  ngOnInit(): void {
    this.projectIdentify = +this.route.snapshot.parent.params['id'];

    this.proService.listProjects("", undefined).toPromise()
        .then(response => {
          let projects = response.json() as Project[];
          if (projects.length) {
            let project = projects.find(data => data.project_id === this.projectIdentify);
            if (project) {
              this.projectName = project.name;
            }
          }
        });
  }

  public get isSystemAdmin(): boolean {
    let account = this.session.getCurrentUser();
    return account != null && account.has_admin_role;
  }

  ngAfterViewInit(): void {
    let isCreated: boolean = this.route.snapshot.queryParams['is_create'];
    if (isCreated) {
      if (this.replicationView) {
        this.replicationView.openModal();
      }
    }
  }

  goRegistry(): void {
    this.router.navigate(['/harbor', 'registries']);
  }
}
