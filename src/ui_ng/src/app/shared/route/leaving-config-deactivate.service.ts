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
import { Injectable } from '@angular/core';
import {
  CanDeactivate, Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot
} from '@angular/router';

import { ConfirmationDialogService } from '../confirmation-dialog/confirmation-dialog.service';

import { ConfigurationComponent } from '../../config/config.component';
import { ConfirmationMessage } from '../confirmation-dialog/confirmation-message';
import { ConfirmationState, ConfirmationTargets } from '../shared.const';

@Injectable()
export class LeavingConfigRouteDeactivate implements CanDeactivate<ConfigurationComponent> {
  constructor(
    private router: Router,
    private confirmation: ConfirmationDialogService) { }

  canDeactivate(
    config: ConfigurationComponent,
    route: ActivatedRouteSnapshot,
    state: RouterStateSnapshot): Promise<boolean> | boolean {
    //Confirmation before leaving config route
    return new Promise((resolve, reject) => {
      if (config && config.hasChanges()) {
        let msg: ConfirmationMessage = new ConfirmationMessage(
          "CONFIG.LEAVING_CONFIRMATION_TITLE",
          "CONFIG.LEAVING_CONFIRMATION_SUMMARY",
          '',
          {},
          ConfirmationTargets.CONFIG_ROUTE
        );
        this.confirmation.openComfirmDialog(msg);
        return this.confirmation.confirmationConfirm$.subscribe(msg => {
          if (msg && msg.source === ConfirmationTargets.CONFIG_ROUTE) {
            if (msg.state === ConfirmationState.CONFIRMED) {
              return resolve(true);
            } else {
              return resolve(false);//Prevent leading route
            }
          } else {
            return resolve(true);//Should go on
          }
        });
      } else {
        return resolve(true);
      }
    });
  }
}
