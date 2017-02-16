import { Component, Output, EventEmitter } from '@angular/core';

export const projectTypes = [
  { 'key' : 0, 'value': 'My Projects' },
  { 'key' : 1, 'value': 'Public Projects'}
];

@Component({
  selector: 'filter-project',
  templateUrl: 'filter-project.component.html'
})
export class FilterProjectComponent {

  @Output() filter = new EventEmitter<number>();
  types = projectTypes;
  currentType = projectTypes[0];

  doFilter(type: number) {
    console.log('Filtered projects by:' + type);
    this.currentType = projectTypes.find(item=>item.key === type);
    this.filter.emit(type);
  }
}