import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

@Component({
  selector: 'tag-repository',
  templateUrl: 'tag-repository.component.html'
})
export class TagRepositoryComponent implements OnInit {

  projectId: number;
  repoName: string;

  constructor(private route: ActivatedRoute) {}
  
  ngOnInit() {
    this.projectId = this.route.snapshot.params['id'];  
    this.repoName = this.route.snapshot.params['repo'];
  }

  deleteTag(tagName: string) {
    
  }

}