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
export class LeavingArtifactSummaryRouteDeactivate implements CanDeactivate<ArtifactListPageComponent> {
  constructor(
    private router: Router,
    private confirmation: ConfirmationDialogService) { }

  canDeactivate(
    tagRepo: ArtifactListPageComponent,
    route: ActivatedRouteSnapshot,
    state: RouterStateSnapshot): Observable<boolean> | boolean {
    // Confirmation before leaving config route
    return new Observable((observer) => {
        sessionStorage.removeItem('referenceSummary');

        return observer.next(true);
    });
  }
}
