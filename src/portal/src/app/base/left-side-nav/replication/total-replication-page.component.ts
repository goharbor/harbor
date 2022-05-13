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
import { Component, OnDestroy, OnInit } from '@angular/core';
import { Router, ActivatedRoute, NavigationEnd } from '@angular/router';
import { SessionService } from '../../../shared/services/session.service';
import { AppConfigService } from '../../../services/app-config.service';
import { Subscription } from 'rxjs';
import {
    EventService,
    HarborEvent,
} from '../../../services/event-service/event.service';
// The route path which will display this component
const URL_TO_DISPLAY: string = '/harbor/replications';
@Component({
    selector: 'total-replication',
    templateUrl: 'total-replication-page.component.html',
    styleUrls: ['./total-replication-page.component.scss'],
})
export class TotalReplicationPageComponent implements OnInit, OnDestroy {
    routerSub: Subscription;
    scrollSub: Subscription;
    scrollTop: number;
    constructor(
        private router: Router,
        private session: SessionService,
        private appConfigService: AppConfigService,
        private activeRoute: ActivatedRoute,
        private event: EventService
    ) {}
    ngOnInit(): void {
        if (!this.scrollSub) {
            this.scrollSub = this.event.subscribe(HarborEvent.SCROLL, v => {
                if (v && URL_TO_DISPLAY === v.url) {
                    this.scrollTop = v.scrollTop;
                }
            });
        }
        if (!this.routerSub) {
            this.routerSub = this.router.events.subscribe(e => {
                if (e instanceof NavigationEnd) {
                    if (e && URL_TO_DISPLAY === e.url) {
                        // Into view
                        this.event.publish(
                            HarborEvent.SCROLL_TO_POSITION,
                            this.scrollTop
                        );
                    } else {
                        this.event.publish(HarborEvent.SCROLL_TO_POSITION, 0);
                    }
                }
            });
        }
    }
    ngOnDestroy(): void {
        if (this.routerSub) {
            this.routerSub.unsubscribe();
            this.routerSub = null;
        }
        if (this.scrollSub) {
            this.scrollSub.unsubscribe();
            this.scrollSub = null;
        }
    }
    goRegistry(): void {
        this.router.navigate(['harbor', 'registries']);
    }

    public get isSystemAdmin(): boolean {
        let account = this.session.getCurrentUser();
        return account != null && account.has_admin_role;
    }

    get withAdmiral(): boolean {
        return this.appConfigService.getConfig().with_admiral;
    }
}
