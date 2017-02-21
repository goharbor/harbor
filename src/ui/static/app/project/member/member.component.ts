import { Component, OnInit } from '@angular/core';
import { Member } from './member';

@Component({
  templateUrl: 'member.component.html'
})
export class MemberComponent implements OnInit {
  members: Member[];

  ngOnInit(): void {
    this.members = [
      { name: 'Admin', role: 'Sys admin'},
      { name: 'user01', role: 'Project Admin'},
      { name: 'user02', role: 'Developer'},
      { name: 'user03', role: 'Guest'}
    ];
  }
}