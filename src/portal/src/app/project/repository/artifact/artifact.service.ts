import { Injectable, Inject } from "@angular/core";
import { Subject } from "rxjs";
import { IServiceConfig, SERVICE_CONFIG } from "../../../../lib/entities/service.config";

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
}
@Injectable()
export class ArtifactDefaultService extends ArtifactService {

  triggerUploadArtifact = new Subject<string>();
  TriggerArtifactChan$ = this.triggerUploadArtifact.asObservable();

  constructor() {
    super();
  }

}
