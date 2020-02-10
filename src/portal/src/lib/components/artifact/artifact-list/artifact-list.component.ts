import { Component, OnInit, Input, OnChanges, SimpleChanges } from '@angular/core';
import { Artifact, Reference } from '../artifact';
import { ArtifactService } from '../../../services';
import { ActivatedRoute } from '@angular/router';
import { forkJoin } from 'rxjs';

@Component({
  selector: 'artifact-list',
  templateUrl: './artifact-list.component.html',
  styleUrls: ['./artifact-list.component.scss']
})
export class ArtifactListComponent implements OnInit, OnChanges {
  @Input() artifactMainifest: Artifact = new Artifact('sha2561234');
  @Input() paddingLeftIndex: number = 1;
  paddingLeftIndex1: number = this.paddingLeftIndex + 1;
  // hasreferenceIndex = !!this.artifactMainifest.references.length;
  referenceNameOpenState = false;
  referenceIndexOpenState = false;
  referenceDigestOpenState = false;
  repoName: string;
  referenceArtifactList: Artifact[] = [];
  hasReferenceArtifactList: Artifact[] = [];
  noReferenceArtifactList: Artifact[] = [];
  @Input() projectName: string;
  constructor(
    public route: ActivatedRoute,
    public artifactService: ArtifactService
  ) {
    this.paddingLeftIndex = ++this.paddingLeftIndex;
    this.paddingLeftIndex1 = this.paddingLeftIndex + 1;


  }

  ngOnInit() {
    this.repoName = this.route.snapshot.params['repo'];
  }
  ngOnChanges(changes: SimpleChanges): void {
    if (changes && changes["paddingLeftIndex"]) {
      this.paddingLeftIndex = this.paddingLeftIndex++;
      this.paddingLeftIndex1 = this.paddingLeftIndex + 1;

    }
  }

  // openArtifact(references: Reference[], indexOrDigest: string) {
  openArtifact(references: Reference[]) {
      
      if (this.referenceNameOpenState === true) {
        this.referenceNameOpenState = false;
        this.referenceIndexOpenState = false;
        this.referenceDigestOpenState = false;
        return;
      }
      if (this.noReferenceArtifactList.length || this.hasReferenceArtifactList.length) {
        this.referenceNameOpenState = true;
        return;
      }
    // this.getArtifactListFromReference(references, indexOrDigest);

  }
  openArtifactContent(indexOrDigest: string) {
    if (indexOrDigest === 'index') {
      if (this.referenceIndexOpenState) {
        this.referenceIndexOpenState = false;
        this.referenceDigestOpenState = false;
        return;
      }

      if (this.hasReferenceArtifactList.length) {
        this.referenceIndexOpenState = true;
        return;
      }
    }
    if (indexOrDigest === 'digest') {
      if (this.referenceDigestOpenState) {
        this.referenceDigestOpenState = false;
        return;
      }
      if (this.noReferenceArtifactList.length) {
        this.referenceDigestOpenState = true;
        return;
      }
    }
    // this.getArtifactListFromReference(references, indexOrDigest);
  }
  // getArtifactListFromReference(references: Reference[], indexOrDigest) {
  //   let artifactObList =
  //     references.map(reference => this.artifactService.getArtifactFromId(this.projectName, this.repoName, reference.artifact_id));
  //   console.log(artifactObList);
  //   forkJoin(artifactObList).subscribe(artifactList => {
  //     // indexOrDigest === 'index' ? this.referenceIndexOpenState = true : this.referenceDigestOpenState = true;
  //     // this.referenceArtifactList = artifactList;
  //     this.referenceNameOpenState = true;
  //     artifactList.forEach(artifact => {
  //       if (artifact.references.length) {
  //         this.hasReferenceArtifactList.push(artifact);
  //       } else {
  //         this.noReferenceArtifactList.push(artifact);
  //       }
  //     });
  //   }, error => {
  //     this.hasReferenceArtifactList.push(new Artifact('sha2560987', 'r'),new Artifact('sha25600000', 't')
  //     );
  //     this.noReferenceArtifactList.push(new Artifact('sha2560123'),new Artifact('sha25600000'),new Artifact('sha25600000')
  //     ,new Artifact('sha25600000'),new Artifact('sha25600000'));
  //     // this.referenceArtifactList = [new Artifact('sha2560987')];
  //     this.referenceNameOpenState = true;

  //     // indexOrDigest === 'index' ? this.referenceIndexOpenState = true : this.referenceDigestOpenState = true;
  //   });
  // }
  openArtifactIndex() {

  }
}
