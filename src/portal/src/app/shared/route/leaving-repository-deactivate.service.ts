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
import { Injectable } from '@angular/core';
import {
  CanDeactivate, Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot
} from '@angular/router';

import { ConfirmationDialogService } from '../confirmation-dialog/confirmation-dialog.service';

import { ConfirmationMessage } from '../confirmation-dialog/confirmation-message';
import { ConfirmationState, ConfirmationTargets } from '../shared.const';
import { ArtifactListPageComponent } from '../../repository/artifact-list-page/artifact-list-page.component';
import { Observable } from 'rxjs';

@Injectable()
export class LeavingRepositoryRouteDeactivate implements CanDeactivate<ArtifactListPageComponent> {
  constructor(
    private router: Router,
    private confirmation: ConfirmationDialogService) { }

  canDeactivate(
    tagRepo: ArtifactListPageComponent,
    route: ActivatedRouteSnapshot,
    state: RouterStateSnapshot): Observable<boolean> | boolean {
    // Confirmation before leaving config route
    return new Observable((observer) => {
      if (tagRepo && tagRepo.hasChanges()) {
        let msg: ConfirmationMessage = new ConfirmationMessage(
          "CONFIG.LEAVING_CONFIRMATION_TITLE",
          "CONFIG.LEAVING_CONFIRMATION_SUMMARY",
          '',
          {},
          ConfirmationTargets.REPOSITORY
        );
        this.confirmation.openComfirmDialog(msg);
        return this.confirmation.confirmationConfirm$.subscribe(confirmMsg => {
          if (confirmMsg && confirmMsg.source === ConfirmationTargets.REPOSITORY) {
            if (confirmMsg.state === ConfirmationState.CONFIRMED) {
              return observer.next(true);
            } else {
              return observer.next(false); // Prevent leading route
            }
          } else {
            return observer.next(true); // Should go on
          }
        });
      } else {
        return observer.next(true);
      }
    });
  }
}
