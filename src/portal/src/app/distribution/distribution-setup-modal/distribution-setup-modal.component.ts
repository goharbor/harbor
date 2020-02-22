import { MessageHandlerService } from './../../shared/message-handler/message-handler.service';
import { DistributionService } from './../distribution.service';
import { Component, OnInit, ViewChild } from '@angular/core';
import {
  DistributionInstance,
  DistributionProvider,
  AuthMode,
  AuthModeBasic,
  AuthModeOAuth
} from '../distribution-interface';
import { NgForm } from '@angular/forms';
import { MsgChannelService } from '../msg-channel.service';
import { TranslateService } from '@ngx-translate/core';
import { errorHandler } from '../../../lib/utils/shared/shared.utils';

@Component({
  selector: 'dist-setup-modal',
  templateUrl: './distribution-setup-modal.component.html',
  styleUrls: ['./distribution-setup-modal.component.scss']
})
export class DistributionSetupModalComponent implements OnInit {
  providers: DistributionProvider[];
  model: DistributionInstance;
  opened: boolean = false;
  editingMode: boolean = false;
  basicUsername: string;
  basicPassword: string;
  authToken: string;
  authData: AuthModeBasic | AuthModeOAuth;
  @ViewChild('instanceForm', { static: true }) instanceForm: NgForm;

  constructor(
    private distributionService: DistributionService,
    private msgHandler: MessageHandlerService,
    private chanService: MsgChannelService,
    private translate: TranslateService
  ) {}

  ngOnInit() {
    this.distributionService.getProviderDrivers().subscribe(
      providers => (this.providers = providers),
      err => console.error(err)
    );
    this.reset();
  }

  public get isValid(): boolean {
    return this.instanceForm && this.instanceForm.valid;
  }

  get title(): string {
    return this.editingMode
      ? 'DISTRIBUTION.EDIT_INSTANCE'
      : 'DISTRIBUTION.SETUP_NEW_INSTANCE';
  }

  authModeChange() {
    switch (this.model.auth_mode) {
      case AuthMode.BASIC:
        this.authData = {
          password: '',
          username: ''
        };
        break;
      case AuthMode.OAUTH:
        this.authData = {
          token: ''
        };
        break;

      default:
        this.authData = null;
        break;
    }
  }

  _open() {
    this.opened = true;
  }

  _close() {
    this.opened = false;
    this.reset();
  }

  reset() {
    this.model = {
      name: '',
      endpoint: '',
      enabled: true,
      setup_timestamp: 0,
      provider: '',
      auth_mode: AuthMode.NONE,
      auth_data: this.authData
    };
    this.instanceForm.reset();
  }

  _isInstance(obj: any): obj is DistributionInstance {
    return obj.endpoint !== undefined;
  }

  cancel() {
    this._close();
  }

  submit() {
    this.model.setup_timestamp = Math.round(new Date().getTime() / 1000);
    if (this.editingMode) {
      this.distributionService.updateInstance(this.model).subscribe(
        response => {
          this.translate.get('DISTRIBUTION.UPDATE_SUCCESS').subscribe(msg => {
            this.msgHandler.info(msg);
          });
          this.chanService.publish('updated');
        },
        err => {
          const message = errorHandler(err);
          this.translate.get('DISTRIBUTION.UPDATE_FAILED').subscribe(msg => {
            this.translate.get(message).subscribe(errMsg => {
              this.msgHandler.error(msg + ': ' + errMsg);
            });
          });
        }
      );
    } else {
      this.distributionService.createInstance(this.model).subscribe(
        response => {
          this.translate.get('DISTRIBUTION.CREATE_SUCCESS').subscribe(msg => {
            this.msgHandler.info(msg);
          });
          this.chanService.publish('created');
        },
        err => {
          const message = errorHandler(err);
          this.translate.get('DISTRIBUTION.CREATE_FAILED').subscribe(msg => {
            this.translate.get(message).subscribe(errMsg => {
              this.msgHandler.error(msg + ': ' + errMsg);
            });
          });
        }
      );
    }

    this._close();
  }

  openSetupModal(editingMode: boolean, data?: DistributionInstance): void {
    this.editingMode = editingMode;
    this._open();

    if (editingMode) {
      this.model = data;
      return;
    }
  }
}
