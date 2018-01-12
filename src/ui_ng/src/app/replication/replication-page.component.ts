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

@Component({
  selector: 'replicaton',
  templateUrl: 'replication-page.component.html'
})
export class ReplicationPageComponent implements OnInit, AfterViewInit {
  projectIdentify: string | number;
  @ViewChild("replicationView") replicationView: ReplicationComponent;

  constructor(private route: ActivatedRoute,
              private router: Router,
              private session: SessionService) { }

  ngOnInit(): void {
    this.projectIdentify = +this.route.snapshot.parent.params['id'];
  }

  public get isSystemAdmin(): boolean {
    let account = this.session.getCurrentUser();
    return account != null && account.has_admin_role > 0;
  }

  openEditPage(id: number): void {
    this.router.navigate(['harbor', 'replications', id, 'rule', { projectId: this.projectIdentify}]);
  }

  openCreatePage(): void {
    this.router.navigate(['harbor', 'replications', 'new-rule', { projectId: this.projectIdentify}] );
  }

  ngAfterViewInit(): void {
    let isCreated: boolean = this.route.snapshot.queryParams['is_create'];
    if (isCreated) {
      if (this.replicationView) {
        this.replicationView.openModal();
      }
    }
  }
}