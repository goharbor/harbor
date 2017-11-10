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
import { Component, Input, Output, EventEmitter, ViewChild, AfterViewChecked } from '@angular/core';

import { NgForm } from '@angular/forms';

import { ReplicationService } from '../service/replication.service';
import { EndpointService } from '../service/endpoint.service';

import { ErrorHandler } from '../error-handler/error-handler';
import { ActionType } from '../shared/shared.const';

import { InlineAlertComponent } from '../inline-alert/inline-alert.component';

import { ReplicationRule } from '../service/interface';
import { Endpoint } from '../service/interface';

import { TranslateService } from '@ngx-translate/core';

import { CREATE_EDIT_RULE_STYLE } from './create-edit-rule.component.css';
import { CREATE_EDIT_RULE_TEMPLATE } from './create-edit-rule.component.html';

import { toPromise } from '../utils';

/**
 * Rule form model.
 */
export interface CreateEditRule {
  ruleId?: number | string;
  name?: string;
  description?: string;
  enable?: boolean;
  endpointId?: number | string;
  endpointName?: string;
  endpointUrl?: string;
  username?: string;
  password?: string;
  insecure?: boolean;
}

const FAKE_PASSWORD: string = 'ywJZnDTM';

@Component({
  selector: 'create-edit-rule',
  template: CREATE_EDIT_RULE_TEMPLATE,
  styles: [ CREATE_EDIT_RULE_STYLE ]
})
export class CreateEditRuleComponent implements AfterViewChecked {

  modalTitle: string;
  createEditRuleOpened: boolean;
  createEditRule: CreateEditRule = this.initCreateEditRule;
  initVal: CreateEditRule = this.initCreateEditRule;
  
  actionType: ActionType;
  
  isCreateEndpoint: boolean;
  @Input() projectId: number;

  @Output() reload = new EventEmitter();

  endpoints: Endpoint[];
  
  pingTestMessage: string;
  testOngoing: boolean;
  pingStatus: boolean;

  btnAbled:boolean;


  ruleForm: NgForm;

  staticBackdrop: boolean = true;
  closable: boolean = false;

  @ViewChild('ruleForm')
  currentForm: NgForm;

  hasChanged: boolean;

  editable: boolean;

  get initCreateEditRule(): CreateEditRule {
    return {
      endpointId: '',
      name: '',
      enable: false,
      description: '',
      endpointName: '',
      endpointUrl: '',
      username: '',
      password: '',
      insecure: false
    };
  }

  get initReplicationRule(): ReplicationRule {
    return {
      project_id: '',
      project_name: '',
      target_id: '',
      target_name: '',
      enabled: 0,
      description: '',
      cron_str: '',
      error_job_count: 0,
      deleted: 0
    };
  }

  get initEndpoint(): Endpoint {
    return {
      endpoint: '',
      name: '',
      username: '',
      password: '',
      insecure: false,
      type: 0
    };
  }

  @ViewChild(InlineAlertComponent)
  inlineAlert: InlineAlertComponent;

  get readonly(): boolean {
    return this.actionType === ActionType.EDIT && (this.createEditRule.enable || false);
  }

  get untoggleable(): boolean {
    return this.actionType === ActionType.EDIT && (this.initVal.enable || false);
  }


  get showNewDestination(): boolean {
    return this.actionType === ActionType.ADD_NEW || (!this.createEditRule.enable || false);
  }
  get connectAbled():boolean{
    return !this.createEditRule.endpointId &&  !this.isCreateEndpoint;

  }

  constructor(
    private replicationService: ReplicationService,
    private endpointService: EndpointService,
    private errorHandler: ErrorHandler,
    private translateService: TranslateService) {}
  
  prepareTargets(endpointId?: number | string) {
    toPromise<Endpoint[]>(this.endpointService
        .getEndpoints())
        .then(endpoints=>{
            this.endpoints = endpoints; 
            if(this.endpoints && this.endpoints.length > 0) {
              let initialEndpoint: Endpoint | undefined;
              (endpointId) ? initialEndpoint = this.endpoints.find(t=>t.id===endpointId) : initialEndpoint = this.endpoints[0]; 
              if(!initialEndpoint) {
                return;
              } 
              this.createEditRule.endpointId = initialEndpoint.id;
              this.createEditRule.endpointName = initialEndpoint.name;
              this.createEditRule.endpointUrl = initialEndpoint.endpoint;
              this.createEditRule.username = initialEndpoint.username;
              this.createEditRule.insecure = initialEndpoint.insecure;
              this.createEditRule.password = FAKE_PASSWORD;

              this.initVal.endpointId = this.createEditRule.endpointId;
              this.initVal.endpointUrl = this.createEditRule.endpointUrl;
              this.initVal.username = this.createEditRule.username;
              this.initVal.password = this.createEditRule.password;
              this.initVal.insecure = this.createEditRule.insecure;
            }
          })
          .catch(error=>{ 
            this.errorHandler.error(error);
            this.createEditRuleOpened = false;
          });
  }

  openCreateEditRule(editable: boolean, ruleId?: number | string): void {

    this.createEditRule = this.initCreateEditRule;
    this.editable = editable;

    this.isCreateEndpoint = false;
    this.hasChanged = false;

    this.pingTestMessage = '';
    this.pingStatus = true;
    this.testOngoing = false;  

    if(ruleId) {
      this.actionType = ActionType.EDIT;
      this.translateService.get('REPLICATION.EDIT_POLICY_TITLE').subscribe(res=>this.modalTitle=res);
      toPromise<ReplicationRule>(this.replicationService
          .getReplicationRule(ruleId))
          .then(rule=>{
            if(rule) {
              this.createEditRule.ruleId = ruleId;
              this.createEditRule.name = rule.name;
              this.createEditRule.description = rule.description;
              this.createEditRule.enable = rule.enabled === 1? true : false;
              this.prepareTargets(rule.target_id);         

              this.initVal.name = this.createEditRule.name;
              this.initVal.description = this.createEditRule.description;
              this.initVal.enable = this.createEditRule.enable;

              this.createEditRuleOpened = true;
            }
          }).catch(err=>this.errorHandler.error(err));
    } else {
      if(!this.projectId) {
        this.errorHandler.error('Project ID cannot be unset');
        return;
      }
      this.actionType = ActionType.ADD_NEW;
      this.translateService.get('REPLICATION.ADD_POLICY').subscribe(res=>this.modalTitle=res);
      this.prepareTargets(); 
      this.createEditRuleOpened = true;
    }
  } 

  newEndpoint(checkedAddNew: boolean): void {
    this.isCreateEndpoint = checkedAddNew;
    if(this.isCreateEndpoint) {
      this.createEditRule.endpointName = '';
      this.createEditRule.endpointUrl = '';
      this.createEditRule.username = '';
      this.createEditRule.password = '';
      this.createEditRule.insecure = false;
    } else {
      this.prepareTargets();
    }
  }

  selectEndpoint(): void {
    let result: Endpoint | undefined = this.endpoints.find(target=>target.id == this.createEditRule.endpointId);
    if(result) {
      this.createEditRule.endpointId = result.id;
      this.createEditRule.endpointUrl = result.endpoint;
      this.createEditRule.username = result.username;
      this.createEditRule.insecure = result.insecure;
      this.createEditRule.password = FAKE_PASSWORD;
    }
  }
    
  getRuleByForm(): ReplicationRule {
    let rule: ReplicationRule = this.initReplicationRule;
    rule.project_id = this.projectId;
    rule.id = this.createEditRule.ruleId;
    rule.name = this.createEditRule.name;
    rule.description = this.createEditRule.description;
    rule.enabled = this.createEditRule.enable ? 1 : 0;
    rule.target_id = this.createEditRule.endpointId || '';
    return rule;
  }

  getEndpointByForm(): Endpoint {
    let endpoint: Endpoint = this.initEndpoint;
    endpoint.id = this.createEditRule.ruleId;
    endpoint.name = this.createEditRule.endpointName || '';
    endpoint.endpoint = this.createEditRule.endpointUrl || '';
    endpoint.username = this.createEditRule.username;
    endpoint.password = this.createEditRule.password;
    endpoint.insecure = this.createEditRule.insecure;
    return endpoint;
  }

  createReplicationRule(): void {
    toPromise<ReplicationRule>(this.replicationService
        .createReplicationRule(this.getRuleByForm()))
        .then(response=>{
            this.translateService.get('REPLICATION.CREATED_SUCCESS')
                .subscribe(res=>this.errorHandler.info(res));
            this.createEditRuleOpened = false;
            this.reload.emit(true);
          })
        .catch(error=>{
            if (error.status === 409) {
              this.inlineAlert.showInlineError('REPLICATION.POLICY_ALREADY_EXISTS');
            } else {
              this.inlineAlert.showInlineError(error);
            }            
          });
  }

  updateReplicationRule(): void {
    toPromise<ReplicationRule>(this.replicationService
        .updateReplicationRule(this.getRuleByForm()))
        .then(()=>{
            this.translateService.get('REPLICATION.UPDATED_SUCCESS')
                .subscribe(res=>this.errorHandler.info(res));
            this.createEditRuleOpened = false;
            this.reload.emit(true);
          })
        .catch(error=>{
            if (error.status === 409) {
              this.inlineAlert.showInlineError('REPLICATION.POLICY_ALREADY_EXISTS');
            } else {
              this.inlineAlert.showInlineError(error);
            }
          }
        );
  }

  createWithEndpoint(actionType: ActionType): void {
    toPromise<Endpoint>(this.endpointService
      .createEndpoint(this.getEndpointByForm()))
      .then(()=>{
        toPromise<Endpoint[]>(this.endpointService
          .getEndpoints(this.createEditRule.endpointName))
          .then(endpoints=>{
            if(endpoints && endpoints.length > 0) {
              let addedEndpoint: Endpoint = endpoints[0];
              this.createEditRule.endpointId = addedEndpoint.id;
              switch(actionType) {
              case ActionType.ADD_NEW:
                this.createReplicationRule();
                break;
              case ActionType.EDIT:
                this.updateReplicationRule();
                break;
              }
            }
         })
         .catch(error=>{
           this.inlineAlert.showInlineError(error);
           this.errorHandler.error(error);
         });
      })
      .catch(error=>{
        this.inlineAlert.showInlineError(error);
        this.errorHandler.error(error);
      });
  }

  onSubmit() {
    if(this.isCreateEndpoint) {
      this.createWithEndpoint(this.actionType);
    } else {
      switch(this.actionType) {
      case ActionType.ADD_NEW:
        this.createReplicationRule();
        break;
      case ActionType.EDIT:
        this.updateReplicationRule();
        break;
      }
    }
  }

  onCancel() {
    if(this.hasChanged) {
      this.inlineAlert.showInlineConfirmation({message: 'ALERT.FORM_CHANGE_CONFIRMATION'});
    } else {
      this.createEditRuleOpened = false;
      this.ruleForm.reset();
    }
  }

  setInsecureValue($event: any) {
    this.createEditRule.insecure = !$event;
  }

  confirmCancel(confirmed: boolean) {
    this.createEditRuleOpened = false;
    this.inlineAlert.close();
    this.ruleForm.reset();
  }

  ngAfterViewChecked(): void {
    this.ruleForm = this.currentForm;
    if(this.ruleForm) {
      let comparison: {[key: string]: any} = {
        name: this.initVal.name,
        description: this.initVal.description,
        enable: this.initVal.enable,
        endpointId: this.initVal.endpointId,
        targetName: this.initVal.name,
        endpointUrl: this.initVal.endpointUrl,
        username: this.initVal.username,
        password: this.initVal.password,
        insecure: this.initVal.insecure
      };
      let self: CreateEditRuleComponent | any = this;
      if(self) {
        self.ruleForm.valueChanges.subscribe((data: any)=>{
          for(let key in data) {
            let current = data[key];          
            let origin: string = comparison[key];
            if(((self.actionType === ActionType.EDIT && !self.readonly && !current ) || current) && current !== origin) {
              self.hasChanged = true;
              break;
            } else {
              self.hasChanged = false;
              self.inlineAlert.close();
            }
          }
        });
      }
    }
  }

  testConnection() {
    this.pingStatus = true;
    this.btnAbled=true;
    this.translateService.get('REPLICATION.TESTING_CONNECTION').subscribe(res=>this.pingTestMessage=res);
    this.testOngoing = !this.testOngoing;
    let pingTarget: Endpoint = this.initEndpoint;
    if(this.isCreateEndpoint) {
      pingTarget.endpoint = this.createEditRule.endpointUrl || '';
      pingTarget.username = this.createEditRule.username;
      pingTarget.password = this.createEditRule.password;
      pingTarget.insecure = this.createEditRule.insecure;
    } else {
      pingTarget.id = this.createEditRule.endpointId;
    }
    toPromise<Endpoint>(this.endpointService
        .pingEndpoint(pingTarget))
        .then(()=>{
            this.testOngoing = !this.testOngoing;
            this.translateService.get('REPLICATION.TEST_CONNECTION_SUCCESS').subscribe(res=>this.pingTestMessage=res);
            this.pingStatus = true;
          this.btnAbled=false;
          })
         .catch(error=>{
            this.testOngoing = !this.testOngoing;
            this.translateService.get('REPLICATION.TEST_CONNECTION_FAILURE').subscribe(res=>this.pingTestMessage=res);
            this.pingStatus = false;
           this.btnAbled=false;
          });
  }
}