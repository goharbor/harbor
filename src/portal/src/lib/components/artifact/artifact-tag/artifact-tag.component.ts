
import { Component, OnInit, Input, OnChanges, SimpleChanges } from '@angular/core';
import { Artifact } from '../artifact';
import { Tag } from '../../../services';

@Component({
  selector: 'artifact-tag',
  templateUrl: './artifact-tag.component.html',
  styleUrls: ['./artifact-tag.component.scss']
})
export class ArtifactTagComponent implements OnInit, OnChanges {
  @Input() artifactDetails: Artifact;
  newTagName = {
    name: ''
  };
  isTagNameExist = false;
  newTagformShow = false;
  constructor() { }

  ngOnInit() {
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes && changes["artifactDetails"]) {
        // this.originalConfig = clone(this.currentConfig);
    }
}
addTag() {
  this.newTagformShow = true;

}
cancelAddTag() {
  this.newTagformShow = false;
  this.newTagName = {name: ''};
}
saveAddTag() {
  // api
  this.newTagformShow = false;
  this.newTagName = {name: ''};
}
removeTag(tag: Tag) {

}
existValid() {

}
}
