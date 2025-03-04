// Copyright Project Harbor Authors
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
import { Component, OnInit } from '@angular/core';
import { SessionService } from '../../../shared/services/session.service';
import { ConfigurationService } from '../../../services/config.service';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { finalize } from 'rxjs/operators';
import { Configuration } from '../config/config';

@Component({
    selector: 'app-project-quotas',
    templateUrl: './project-quotas-container.component.html',
    styleUrls: ['./project-quotas-container.component.scss'],
})
export class ProjectQuotasContainerComponent implements OnInit {
    allConfig: Configuration = new Configuration();
    loading: boolean = false;
    constructor(
        private session: SessionService,
        private configService: ConfigurationService,
        private msgHandler: MessageHandlerService
    ) {}

    ngOnInit() {
        let currentUser = this.session.getCurrentUser();
        if (currentUser && currentUser.has_admin_role) {
            this.retrieveConfig();
        }
    }

    refreshAllconfig() {
        this.retrieveConfig();
    }
    retrieveConfig(): void {
        this.loading = true;
        this.configService
            .getConfiguration()
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                (configurations: Configuration) => {
                    this.allConfig = configurations;
                },
                error => {
                    this.msgHandler.handleError(error);
                }
            );
    }
}
