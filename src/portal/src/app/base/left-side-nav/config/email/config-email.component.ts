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
import { Component, OnInit, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { ConfigurationService } from '../../../../services/config.service';
import { Configuration } from '../config';
import {
    isEmpty,
    getChanges as getChangesFunc,
} from '../../../../shared/units/utils';
import { errorHandler } from '../../../../shared/units/shared.utils';
import { ConfigService } from '../config.service';

@Component({
    selector: 'config-email',
    templateUrl: 'config-email.component.html',
    styleUrls: ['./config-email.component.scss', '../config.component.scss'],
})
export class ConfigurationEmailComponent implements OnInit {
    testingMailOnGoing = false;
    onGoing = false;
    @ViewChild('mailConfigFrom', { static: true }) mailForm: NgForm;
    get currentConfig(): Configuration {
        return this.conf.getConfig();
    }

    set currentConfig(c: Configuration) {
        this.conf.setConfig(c);
    }

    constructor(
        private msgHandler: MessageHandlerService,
        private configService: ConfigurationService,
        private conf: ConfigService
    ) {}

    ngOnInit(): void {
        this.conf.resetConfig();
    }

    disabled(prop: any): boolean {
        return !(prop && prop.editable);
    }

    setInsecureValue($event: any) {
        this.currentConfig.email_insecure.value = !$event;
    }

    isValid(): boolean {
        return this.mailForm && this.mailForm.valid;
    }

    inProgress(): boolean {
        return this.onGoing || this.conf.getLoadingConfigStatus();
    }

    public hasChanges(): boolean {
        return !isEmpty(this.getChanges());
    }

    public getChanges() {
        let allChanges = getChangesFunc(
            this.conf.getOriginalConfig(),
            this.currentConfig
        );
        let changes = {};
        for (let prop in allChanges) {
            if (prop.startsWith('email_')) {
                changes[prop] = allChanges[prop];
            }
        }
        return changes;
    }
    /**
     *
     * Test the connection of specified mail server
     *
     *
     * @memberOf ConfigurationComponent
     */
    public testMailServer(): void {
        if (this.testingMailOnGoing) {
            return; // Should not come here
        }
        let mailSettings = {};
        for (let prop in this.currentConfig) {
            if (prop.startsWith('email_')) {
                mailSettings[prop] = this.currentConfig[prop].value;
            }
        }
        // Confirm port is number
        mailSettings['email_port'] = +mailSettings['email_port'];
        let allChanges = this.getChanges();
        let password = allChanges['email_password'];
        if (password) {
            mailSettings['email_password'] = password;
        } else {
            delete mailSettings['email_password'];
        }

        this.testingMailOnGoing = true;
        this.configService.testMailServer(mailSettings).subscribe(
            response => {
                this.testingMailOnGoing = false;
                this.msgHandler.showSuccess('CONFIG.TEST_MAIL_SUCCESS');
            },
            error => {
                this.testingMailOnGoing = false;
                let err = errorHandler(error);
                if (!err) {
                    err = 'UNKNOWN';
                }
                this.msgHandler.showError('CONFIG.TEST_MAIL_FAILED', {
                    param: err,
                });
            }
        );
    }

    public get hideMailTestingSpinner(): boolean {
        return !this.testingMailOnGoing;
    }

    public isMailConfigValid(): boolean {
        return this.isValid() && !this.testingMailOnGoing && !this.inProgress();
    }

    /**
     *
     * Save the changed values
     *
     * @memberOf ConfigurationComponent
     */
    public save(): void {
        let changes = this.getChanges();
        if (!isEmpty(changes)) {
            this.onGoing = true;
            this.configService.saveConfiguration(changes).subscribe(
                response => {
                    this.onGoing = false;
                    // refresh allConfig
                    this.conf.updateConfig();
                    this.msgHandler.showSuccess('CONFIG.SAVE_SUCCESS');
                },
                error => {
                    this.onGoing = false;
                    this.msgHandler.handleError(error);
                }
            );
        } else {
            // Inprop situation, should not come here
            console.error('Save abort because nothing changed');
        }
    }

    /**
     *
     * Discard current changes if have and reset
     *
     * @memberOf ConfigurationComponent
     */
    public cancel(): void {
        let changes = this.getChanges();
        if (!isEmpty(changes)) {
            this.conf.confirmUnsavedChanges(changes);
        } else {
            // Invalid situation, should not come here
            console.error('Nothing changed');
        }
    }
}
