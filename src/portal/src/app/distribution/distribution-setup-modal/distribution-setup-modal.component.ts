import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { Component, OnInit, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';
import { MsgChannelService } from '../msg-channel.service';
import { TranslateService } from '@ngx-translate/core';
import { errorHandler } from '../../../lib/utils/shared/shared.utils';
import { PreheatService } from "../../../../ng-swagger-gen/services/preheat.service";
import { Instance } from "../../../../ng-swagger-gen/models/instance";
import { Provider } from "../../../../ng-swagger-gen/models/provider";
import { AuthMode } from "../distribution-interface";
import { clone } from "../../../lib/utils/utils";
import { InlineAlertComponent } from "../../shared/inline-alert/inline-alert.component";
import { ClrLoadingState } from "@clr/angular";

@Component({
  selector: 'dist-setup-modal',
  templateUrl: './distribution-setup-modal.component.html',
  styleUrls: ['./distribution-setup-modal.component.scss']
})
export class DistributionSetupModalComponent implements OnInit {
  providers: Provider[];
  model: Instance;
  opened: boolean = false;
  editingMode: boolean = false;
  basicUsername: string;
  basicPassword: string;
  authToken: string;
  authData: {[key: string]: any};
  @ViewChild('instanceForm', { static: true }) instanceForm: NgForm;
  @ViewChild(InlineAlertComponent, { static: false }) inlineAlert: InlineAlertComponent;
  saveBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;

  constructor(
    private distributionService: PreheatService,
    private msgHandler: MessageHandlerService,
    private chanService: MsgChannelService,
    private translate: TranslateService
  ) {}

  ngOnInit() {
    this.distributionService.ListProviders().subscribe(
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
    this.inlineAlert.close();
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
      provider: '',
      auth_mode: AuthMode.NONE,
      auth_data: this.authData
    };
    this.instanceForm.reset();
  }

  _isInstance(obj: any): obj is Instance {
    return obj.endpoint !== undefined;
  }

  cancel() {
    this._close();
  }

  submit() {
    if (this.editingMode) {
      const data: Instance = {
        endpoint: this.model.endpoint,
        enabled: this.model.enabled,
        description: this.model.description,
        auth_mode: this.model.auth_mode,
        auth_data: this.model.auth_data
      };
      this.saveBtnState = ClrLoadingState.LOADING;
      this.distributionService.UpdateInstance({instanceId: this.model.id, propertySet: data
        }).subscribe(
        response => {
          this.translate.get('DISTRIBUTION.UPDATE_SUCCESS').subscribe(msg => {
            this.msgHandler.info(msg);
          });
          this.chanService.publish('updated');
          this.saveBtnState = ClrLoadingState.SUCCESS;
          this._close();
        },
        err => {
          const message = errorHandler(err);
          this.translate.get('DISTRIBUTION.UPDATE_FAILED').subscribe(msg => {
            this.translate.get(message).subscribe(errMsg => {
              this.inlineAlert.showInlineError(msg + ': ' + errMsg);
              this.saveBtnState = ClrLoadingState.ERROR;
            });
          });
        }
      );
    } else {
      this.saveBtnState = ClrLoadingState.LOADING;
      this.distributionService.CreateInstance({instance: this.model}).subscribe(
        response => {
          this.translate.get('DISTRIBUTION.CREATE_SUCCESS').subscribe(msg => {
            this.msgHandler.info(msg);
          });
          this.chanService.publish('created');
          this.saveBtnState = ClrLoadingState.SUCCESS;
          this._close();
        },
        err => {
          const message = errorHandler(err);
          this.translate.get('DISTRIBUTION.CREATE_FAILED').subscribe(msg => {
            this.translate.get(message).subscribe(errMsg => {
              this.inlineAlert.showInlineError(msg + ': ' + errMsg);
              this.saveBtnState = ClrLoadingState.ERROR;
            });
          });
        }
      );
    }
  }

  openSetupModal(editingMode: boolean, data?: Instance): void {
    this.editingMode = editingMode;
    this._open();
    if (editingMode) {
      this.model = clone(data);
    }
  }
}
