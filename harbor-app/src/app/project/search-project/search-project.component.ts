import { Component, EventEmitter, Output } from '@angular/core';

@Component({
  selector: 'search-project',
  templateUrl: 'search-project.component.html'
})
export class SearchProjectComponent {
  @Output() search = new EventEmitter<string>();

  doSearch(projectName) {
    this.search.emit(projectName);
  }
}