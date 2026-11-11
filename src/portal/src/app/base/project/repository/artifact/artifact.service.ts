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
import { Observable, Subject } from 'rxjs';
import { DomSanitizer, SafeUrl } from '@angular/platform-browser';
import { Artifact } from '../../../../../../ng-swagger-gen/models/artifact';
import { IconService } from '../../../../../../ng-swagger-gen/services/icon.service';
import { share } from 'rxjs/operators';
import { Icon } from 'ng-swagger-gen/models/icon';
import { Accessory } from '../../../../../../ng-swagger-gen/models/accessory';
import {
    EventService,
    HarborEvent,
} from '../../../../services/event-service/event.service';

/**
 * Define the service methods to handle the repository tag related things.
 *
 **
 * @abstract
 * class ArtifactService
 */
export abstract class ArtifactService {
    reference: string[];
    triggerUploadArtifact = new Subject<string>();
    abstract getIcon(digest: string): SafeUrl;
    abstract setIcon(digest: string, url: SafeUrl);
    abstract getIconsFromBackEnd(artifactList: Artifact[] | Accessory[]);
}
@Injectable()
export class ArtifactDefaultService extends ArtifactService {
    triggerUploadArtifact = new Subject<string>();
    private _iconMap: { [key: string]: SafeUrl } = {};
    private _sharedIconObservableMap: { [key: string]: Observable<Icon> } = {};
    constructor(
        private iconService: IconService,
        private domSanitizer: DomSanitizer,
        private event: EventService
    ) {
        super();
    }
    getIcon(icon: string): SafeUrl {
        return this._iconMap[icon];
    }
    setIcon(icon: string, url: SafeUrl) {
        if (!this._iconMap[icon]) {
            this._iconMap[icon] = url;
        }
    }
    getIconsFromBackEnd(artifactList: Artifact[]) {
        if (artifactList && artifactList.length) {
            artifactList.forEach(item => {
                if (item.icon && !this.getIcon(item.icon)) {
                    if (!this._sharedIconObservableMap[item.icon]) {
                        this._sharedIconObservableMap[item.icon] =
                            this.iconService
                                .getIcon({ digest: item.icon })
                                .pipe(share());
                    }
                    this._sharedIconObservableMap[item.icon].subscribe(res => {
                        this.setIcon(
                            item.icon,
                            this.domSanitizer.bypassSecurityTrustUrl(
                                `data:${res['content-type']};charset=utf-8;base64,${res.content}`
                            )
                        );
                        this.event.publish(HarborEvent.RETRIEVED_ICON);
                    });
                }
            });
        }
    }
}
