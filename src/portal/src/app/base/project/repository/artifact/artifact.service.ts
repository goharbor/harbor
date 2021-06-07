import { Injectable } from "@angular/core";
import { Subject } from "rxjs";
import { DomSanitizer, SafeUrl } from '@angular/platform-browser';
import { Artifact } from '../../../../../../ng-swagger-gen/models/artifact';
import { IconService } from '../../../../../../ng-swagger-gen/services/icon.service';


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
  TriggerArtifactChan$ = this.triggerUploadArtifact.asObservable();
  abstract getIcon(digest: string): SafeUrl;
  abstract setIcon(digest: string, url: SafeUrl);
  abstract getIconsFromBackEnd(artifactList: Artifact[]);
}
@Injectable()
export class ArtifactDefaultService extends ArtifactService {

  triggerUploadArtifact = new Subject<string>();
  TriggerArtifactChan$ = this.triggerUploadArtifact.asObservable();
  private _iconMap: {[key: string]: SafeUrl} = {};
  constructor(private iconService: IconService,
              private domSanitizer: DomSanitizer) {
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
          this.iconService.getIcon({digest: item.icon})
            .subscribe(res => {
              this.setIcon(item.icon, this.domSanitizer
                .bypassSecurityTrustUrl(`data:${res['content-type']};charset=utf-8;base64,${res.content}`));
            });
        }
      });
    }
  }
}
