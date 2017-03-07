import { Component, OnInit, OnDestroy, ViewChild } from '@angular/core';
import { Router } from '@angular/router';
import { NgForm } from '@angular/forms';

import { ConfigurationService } from './config.service';
import { Configuration } from './config';
import { MessageService } from '../global-message/message.service';
import { AlertType, DeletionTargets } from '../shared/shared.const';
import { errorHandler, accessErrorHandler } from '../shared/shared.utils';
import { StringValueItem } from './config';
import { DeletionDialogService } from '../shared/deletion-dialog/deletion-dialog.service';
import { Subscription } from 'rxjs/Subscription';
import { DeletionMessage } from '../shared/deletion-dialog/deletion-message'

import { ConfigurationAuthComponent } from './auth/config-auth.component';
import { ConfigurationEmailComponent } from './email/config-email.component';

const fakePass = "fakepassword";

@Component({
    selector: 'config',
    templateUrl: "config.component.html",
    styleUrls: ['config.component.css']
})
export class ConfigurationComponent implements OnInit, OnDestroy {
    private onGoing: boolean = false;
    allConfig: Configuration = new Configuration();
    private currentTabId: string = "";
    private originalCopy: Configuration;
    private confirmSub: Subscription;

    @ViewChild("repoConfigFrom") repoConfigForm: NgForm;
    @ViewChild("systemConfigFrom") systemConfigForm: NgForm;
    @ViewChild(ConfigurationEmailComponent) mailConfig: ConfigurationEmailComponent;
    @ViewChild(ConfigurationAuthComponent) authConfig: ConfigurationAuthComponent;

    constructor(
        private configService: ConfigurationService,
        private msgService: MessageService,
        private confirmService: DeletionDialogService) { }

    ngOnInit(): void {
        //First load
        this.retrieveConfig();

        this.confirmSub = this.confirmService.deletionConfirm$.subscribe(confirmation => {
            this.reset(confirmation.data);
        });
    }

    ngOnDestroy(): void {
        if (this.confirmSub) {
            this.confirmSub.unsubscribe();
        }
    }

    public get inProgress(): boolean {
        return this.onGoing;
    }

    public isValid(): boolean {
        return this.repoConfigForm &&
            this.repoConfigForm.valid &&
            this.systemConfigForm &&
            this.systemConfigForm.valid &&
            this.mailConfig &&
            this.mailConfig.isValid() &&
            this.authConfig &&
            this.authConfig.isValid();
    }

    public hasChanges(): boolean {
        return !this.isEmpty(this.getChanges());
    }

    public isMailConfigValid(): boolean {
        return this.mailConfig &&
            this.mailConfig.isValid();
    }

    public get showTestServerBtn(): boolean {
        return this.currentTabId === 'config-email';
    }

    public tabLinkChanged(tabLink: any) {
        this.currentTabId = tabLink.id;
    }

    /**
     * 
     * Save the changed values
     * 
     * @memberOf ConfigurationComponent
     */
    public save(): void {
        let changes = this.getChanges();
        if (!this.isEmpty(changes)) {
            this.onGoing = true;
            this.configService.saveConfiguration(changes)
                .then(response => {
                    this.onGoing = false;
                    //API should return the updated configurations here
                    //Unfortunately API does not do that
                    //To refresh the view, we can clone the original data copy
                    //or force refresh by calling service.
                    //HERE we choose force way
                    this.retrieveConfig();
                    this.msgService.announceMessage(response.status, "CONFIG.SAVE_SUCCESS", AlertType.SUCCESS);
                })
                .catch(error => {
                    this.onGoing = false;
                    if (!accessErrorHandler(error, this.msgService)) {
                        this.msgService.announceMessage(error.status, errorHandler(error), AlertType.DANGER);
                    }
                });
        } else {
            //Inprop situation, should not come here
            console.error("Save obort becasue nothing changed");
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
        if (!this.isEmpty(changes)) {
            let msg = new DeletionMessage(
                "CONFIG.CONFIRM_TITLE",
                "CONFIG.CONFIRM_SUMMARY",
                "",
                changes,
                DeletionTargets.EMPTY
            );
            this.confirmService.openComfirmDialog(msg);
        } else {
            //Inprop situation, should not come here
            console.error("Nothing changed");
        }
    }

    /**
     * 
     * Test the connection of specified mail server
     * 
     * 
     * @memberOf ConfigurationComponent
     */
    public testMailServer(): void {

    }

    private retrieveConfig(): void {
        this.onGoing = true;
        this.configService.getConfiguration()
            .then(configurations => {
                this.onGoing = false;

                //Add two password fields
                configurations.email_password = new StringValueItem(fakePass, true);
                configurations.ldap_search_password = new StringValueItem(fakePass, true);
                this.allConfig = configurations;

                //Keep the original copy of the data
                this.originalCopy = this.clone(configurations);
            })
            .catch(error => {
                this.onGoing = false;
                if (!accessErrorHandler(error, this.msgService)) {
                    this.msgService.announceMessage(error.status, errorHandler(error), AlertType.DANGER);
                }
            });
    }

    /**
     * 
     * Get the changed fields and return a map
     * 
     * @private
     * @returns {*}
     * 
     * @memberOf ConfigurationComponent
     */
    private getChanges(): any {
        let changes = {};
        if (!this.allConfig || !this.originalCopy) {
            return changes;
        }

        for (let prop in this.allConfig) {
            let field = this.originalCopy[prop];
            if (field && field.editable) {
                if (field.value != this.allConfig[prop].value) {
                    changes[prop] = this.allConfig[prop].value;
                    //Fix boolean issue
                    if (typeof field.value === "boolean") {
                        changes[prop] = changes[prop] ? "1" : "0";
                    }
                }
            }
        }

        return changes;
    }

    /**
     * 
     * Deep clone the configuration object
     * 
     * @private
     * @param {Configuration} src
     * @returns {Configuration}
     * 
     * @memberOf ConfigurationComponent
     */
    private clone(src: Configuration): Configuration {
        let dest = new Configuration();
        if (!src) {
            return dest;//Empty
        }

        for (let prop in src) {
            if (src[prop]) {
                dest[prop] = Object.assign({}, src[prop]); //Deep copy inner object
            }
        }

        return dest;
    }

    /**
     * 
     * Reset the configuration form
     * 
     * @private
     * @param {*} changes
     * 
     * @memberOf ConfigurationComponent
     */
    private reset(changes: any): void {
        if (!this.isEmpty(changes)) {
            for (let prop in changes) {
                if (this.originalCopy[prop]) {
                    this.allConfig[prop] = Object.assign({}, this.originalCopy[prop]);
                }
            }
        } else {
            //force reset
            this.retrieveConfig();
        }
    }

    private isEmpty(obj) {
        for (let key in obj) {
            if (obj.hasOwnProperty(key))
                return false;
        }
        return true;
    }

    private disabled(prop: any): boolean {
        return !(prop && prop.editable);
    }
}